# payments-service-demo

A project that implements a card payment flow with 3D Secure (3DS) using Go.

## Why this exists

This repo is for implementing payment architecture correctly:

- asynchronous payment lifecycles
- 3DS challenge / redirect handling
- webhook-driven reconciliation
- client disconnect recovery
- deployable production-style Go service design

## Provider choice

I'm using Stripe because it offers free test mode, official 3DS test methods, and a mature Go SDK. I will be keeping the application provider-agnostic.

## Local run

```bash
go run ./cmd/api
```

Open http://localhost:3000/healthz.

## Enable pre-commit hook

To enable go fmt to be run pre commit, we need to give the necessary permissions and enable the local hook path before it can be used in development. To do that, run

```bash
chmod +x .githooks/pre-commit
git config core.hooksPath .githooks
```

## Configuration

Environment variables:

- `APP_ENV` - application environment, defaults to `development`
- `APP_VERSION` - runtime application version string, defaults to `dev`
- `PORT` - HTTP port, defaults to `3000`
- `PAYMENTS_PROVIDER` - payment gateway provider, defaults to `fake`
- `STRIPE_SECRET_KEY` - Stripe secret key, used when `PAYMENTS_PROVIDER=stripe`
- `STRIPE_PUBLISHABLE_KEY` - Stripe publishable key, used when `PAYMENTS_PROVIDER=stripe`
- `STRIPE_WEBHOOK_SECRET` - Stripe webhook signing secret, required when `PAYMENTS_PROVIDER=stripe`
- `DATABASE_URL` - PostgreSQL connection string, required

## Available endpoints

- `GET /healthz`
- `POST /payment-attempts`
- `GET /payment-attempts/{id}`
-  `POST /webhooks/stripe`

## Example: create a payment attempt

```bash
curl -i \
  -X POST http://localhost:3000/payment-attempts \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idempotency-key-123" \
  -d '{
    "order_id": "order_123",
    "amount": 2500,
    "currency": "gbp",
    "return_url": "https://example.com/return",
    "description": "test payment"
  }'
```

## Local Postgres setup

This project uses PostgreSQL for durable payment attempt storage.

Start Postgres locally:

```bash
make db-up
```
Apply the schema:

```bash
make db-schema
```

Open a local psql session:

```bash
make db-psql
```

#### Default local connection details:
```
host: localhost
port: 5432
database: payments_service
user: payments_service
password: payments_service
```

Example local DATABASE_URL

```bash
export DATABASE_URL="postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable"
```

The initial schema is in:
```
db/schema.sql
```

## Local webhook testing with Stripe CLI

Install the Stripe CLI and authenticate:
```bash
brew install stripe/stripe-cli/stripe
```
and
```bash
stripe login
```
Start forwarding Stripe webhook events to your local service:
```bash
stripe listen --forward-to localhost:3000/webhooks/stripe
```
The CLI will print a webhook signing secret. Export it before starting the app:
```bash
export STRIPE_WEBHOOK_SECRET="whsec_..."
```
Trigger a test webhook event:
```bash
stripe trigger payment_intent.succeeded
```
You can also trigger other useful events such as:
```bash
stripe trigger payment_intent.payment_failed
stripe trigger payment_intent.processing
```

## Webhook event processing

The service processes these verified Stripe webhook events:

- `payment_intent.succeeded`
- `payment_intent.payment_failed`
- `payment_intent.processing`
- `payment_intent.canceled`

For each supported event, the backend:

1. verifies the Stripe signature using the raw request body
2. parses the webhook event
3. extracts the Stripe `PaymentIntent` ID
4. finds the local `PaymentAttempt` by `provider_payment_id`
5. updates the local payment status in PostgreSQL

Unhandled Stripe event types are currently ignored and acknowledged with `200 OK`.

### Duplicate delivery handling

Stripe may retry webhook deliveries. The service now persists processed Stripe event IDs and ignores duplicate deliveries safely.

On duplicate delivery of the same Stripe event ID:

- the endpoint returns `200 OK`
- the event is not reprocessed
- local payment state is not updated a second time

## Dependencies
This project currently uses Stripe Go SDK `github.com/stripe/stripe-go/v84`.

## Container build and local run

Build the container image:

```bash
docker build -t payments-service:local .
```

### Run it locally:

```bash
docker run --rm \
  -p 3000:3000 \
  -e PORT=3000 \
  -e DATABASE_URL="postgres://payments_service:payments_service@host.docker.internal:5432/payments_service?sslmode=disable" \
  payments-service:local
```

### Open:

  http://localhost:3000/healthz

  http://localhost:3000/demo

## Cloud deployment

Cloud Run and Cloud SQL deployment notes are in:

```text
deploy/README.md
```

This includes the required secrets, service account roles, and deployment flow.

## Release deployments

This repository supports release-driven Cloud Run deployments through GitHub Actions.

Publishing a GitHub Release such as: `v1.0.0`

will:
- build and push a container image tagged v1.0.0
- deploy that image to Cloud Run
- set APP_VERSION=v1.0.0

See:
```
deploy/github-actions/README.md
```

## Performance testing mode

For performance and load testing, prefer running the service with:

```bash
export PAYMENTS_PROVIDER=fake
```

This avoids using Stripe and allows you to stress-test the backend API, database access, idempotency behavior, and reconciliation flow locally.

## Running local performance tests

refer [k6/README.md](k6/README.md)
