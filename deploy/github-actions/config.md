# GitHub workflow configuration

These workflows are designed to be portable across similar services.

## Required GitHub repository or environment variables

Set these as GitHub Variables:

- `GCP_PROJECT_ID`
- `GCP_REGION`
- `CLOUD_RUN_SERVICE`
- `ARTIFACT_REPOSITORY`
- `IMAGE_NAME`
- `CLOUD_SQL_INSTANCE`
- `APP_ENV`
- `APP_PORT`
- `PAYMENTS_PROVIDER`
- `TEST_DB_NAME`
- `TEST_DB_USER`
- `TEST_DB_PASSWORD`
- `TEST_DB_PORT`
- `MIN_COVERAGE`
- `DATABASE_URL_SECRET_NAME`
- `STRIPE_SECRET_KEY_SECRET_NAME`
- `STRIPE_PUBLISHABLE_KEY_SECRET_NAME`
- `STRIPE_WEBHOOK_SECRET_SECRET_NAME`

## Required GitHub secrets

Set these as GitHub Secrets:

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_DEPLOY_SERVICE_ACCOUNT`
- `GCP_RUNTIME_SERVICE_ACCOUNT`

## Example values for this repository

- `GCP_PROJECT_ID=payments-service-staging`
- `GCP_REGION=us-west1`
- `CLOUD_RUN_SERVICE=payments-service-v1`
- `ARTIFACT_REPOSITORY=payments-service-v1`
- `IMAGE_NAME=payments-service`
- `CLOUD_SQL_INSTANCE=payments-service-staging:us-west1:payments-service-db`
- `APP_ENV=staging`
- `APP_PORT=3000`
- `PAYMENTS_PROVIDER=stripe`
- `TEST_DB_NAME=payments_service`
- `TEST_DB_USER=payments_service`
- `TEST_DB_PASSWORD=payments_service`
- `TEST_DB_PORT=5432`
- `MIN_COVERAGE=70`
- `DATABASE_URL_SECRET_NAME=payments-service-database-url`
- `STRIPE_SECRET_KEY_SECRET_NAME=payments-service-stripe-secret-key`
- `STRIPE_PUBLISHABLE_KEY_SECRET_NAME=payments-service-stripe-publishable-key`
- `STRIPE_WEBHOOK_SECRET_SECRET_NAME=payments-service-stripe-webhook-secret`

## Portability strategy

To reuse these workflows in another service:

1. copy the workflow files
2. set the new repository variables
3. set the required secrets
4. adjust any service-specific runtime secrets if needed
