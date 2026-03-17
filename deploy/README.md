# Cloud Run deployment prerequisites

This service is designed to run on Google Cloud Run with PostgreSQL provided by Cloud SQL.

## Required Google Cloud resources

Before deploying, provision:

- a Google Cloud project
- a Cloud Run service account
- a Cloud SQL for PostgreSQL instance
- a database inside that instance
- Secret Manager secrets for Stripe configuration
- a container image in Artifact Registry

Cloud Run supports:
- container image deployment
- Secret Manager-backed configuration
- Cloud SQL connections for PostgreSQL

## Required IAM roles

At minimum, the Cloud Run service account should have:

- `roles/cloudsql.client`
- `roles/secretmanager.secretAccessor`

Depending on your Cloud SQL authentication strategy, additional database-level setup may be needed.

## Required secrets

Create these Secret Manager secrets:

- `payments-service-stripe-secret-key`
- `payments-service-stripe-publishable-key`
- `payments-service-stripe-webhook-secret`

## Required environment configuration

The service needs these runtime values:

- `APP_ENV=staging`
- `PORT` provided by Cloud Run automatically
- `DATABASE_URL`
- `PAYMENTS_PROVIDER`
- `STRIPE_SECRET_KEY`
- `STRIPE_PUBLISHABLE_KEY`
- `STRIPE_WEBHOOK_SECRET`

For Cloud Run, the service must listen on the injected `PORT` environment variable.
For Cloud SQL PostgreSQL, Cloud Run supports connecting to a Cloud SQL instance as a managed service dependency.

## Database URL note

This service currently expects a standard `DATABASE_URL`.

For Cloud SQL PostgreSQL, you should construct the URL according to your chosen connectivity model and authentication setup.
Keep the final runtime shape as:

```text
postgres://<user>:<password>@<host>:<port>/<db>?sslmode=disable
```

If you use the Cloud SQL Auth Proxy path or Cloud Run's built-in Cloud SQL integration, adapt the host portion accordingly for your environment.


## Deployment command

Make the script executable:

```bash
chmod +x deploy/cloudrun/deploy.sh
```

Set the environment variables from `deploy/cloudrun/env.example`, then run:

```bash
source deploy/cloudrun/env.example
./deploy/cloudrun/deploy.sh
```

After deployment

Run the checks in:
```text
deploy/cloudrun/regression-checklist.md
```
