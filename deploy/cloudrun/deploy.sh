#!/usr/bin/env bash
set -euo pipefail

# Required environment variables:
#   PROJECT_ID
#   REGION
#   SERVICE_NAME
#   IMAGE
#   SERVICE_ACCOUNT
#   CLOUD_SQL_INSTANCE
#   DATABASE_URL
#
# Required Secret Manager secrets:
#   payments-service-stripe-secret-key
#   payments-service-stripe-publishable-key
#   payments-service-stripe-webhook-secret

: "${PROJECT_ID:?PROJECT_ID is required}"
: "${REGION:?REGION is required}"
: "${SERVICE_NAME:?SERVICE_NAME is required}"
: "${IMAGE:?IMAGE is required}"
: "${SERVICE_ACCOUNT:?SERVICE_ACCOUNT is required}"
: "${CLOUD_SQL_INSTANCE:?CLOUD_SQL_INSTANCE is required}"
: "${DATABASE_URL:?DATABASE_URL is required}"

gcloud run deploy "${SERVICE_NAME}" \
  --project="${PROJECT_ID}" \
  --region="${REGION}" \
  --platform=managed \
  --image="${IMAGE}" \
  --service-account="${SERVICE_ACCOUNT}" \
  --allow-unauthenticated \
  --add-cloudsql-instances="${CLOUD_SQL_INSTANCE}" \
  --set-env-vars="APP_ENV=staging,PAYMENTS_PROVIDER=stripe,DATABASE_URL=${DATABASE_URL}" \
  --set-secrets="STRIPE_SECRET_KEY=payments-service-stripe-secret-key:latest,STRIPE_PUBLISHABLE_KEY=payments-service-stripe-publishable-key:latest,STRIPE_WEBHOOK_SECRET=payments-service-stripe-webhook-secret:latest"
