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

Open http://localhost:8080/healthz.

## Enable pre-commit hook

To enable go fmt to be run pre commit, we need to give the necessary permissions and enable the local hook path before it can be used in development. To do that, run

```bash
chmod +x .githooks/pre-commit
git config core.hooksPath .githooks
```

## Configuration

Environment variables:

- `APP_ENV` - application environment, defaults to `development`
- `HTTP_PORT` - HTTP port, defaults to `8080`
- `PAYMENTS_PROVIDER` - payment gateway provider, defaults to `fake`
- `STRIPE_SECRET_KEY` - Stripe secret key, used when `PAYMENTS_PROVIDER=stripe`
- `STRIPE_PUBLISHABLE_KEY` - Stripe publishable key, used when `PAYMENTS_PROVIDER=stripe`
- `DATABASE_URL` - PostgreSQL connection string, required

## Available endpoints

- `GET /healthz`
- `POST /payment-attempts`
- `GET /payment-attempts/{id}`

## Example: create a payment attempt

```bash
curl -i \
  -X POST http://localhost:8080/payment-attempts \
  -H "Content-Type: application/json" \
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
