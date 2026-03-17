# GitHub release deployment setup

This repository can deploy to Cloud Run automatically when a GitHub Release is published.

## Deployment flow

1. Publish a GitHub Release such as `v1.0.0`
2. GitHub Actions authenticates to Google Cloud using OIDC / Workload Identity Federation
3. The workflow builds and pushes a versioned container image to Artifact Registry
4. The workflow deploys that exact image to Cloud Run
5. The workflow sets `APP_VERSION` to the GitHub Release tag

## Why OIDC

Use GitHub OIDC with Google Workload Identity Federation instead of long-lived JSON service account keys.

This avoids storing a permanent GCP credential in GitHub and is the recommended modern pattern.

## Required Google Cloud setup

### 1. Create a deployer service account

Create a service account used by GitHub Actions to deploy:

```bash
gcloud iam service-accounts create github-deployer \
  --display-name="GitHub Actions deployer"
```

Grant it the minimum roles needed for this repository's deployment workflow:

```
gcloud projects add-iam-policy-binding payments-service-staging \
  --member="serviceAccount:github-deployer@payments-service-staging.iam.gserviceaccount.com" \
  --role="roles/run.admin"
```
```
gcloud projects add-iam-policy-binding payments-service-staging \
  --member="serviceAccount:github-deployer@payments-service-staging.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"
```
```
gcloud projects add-iam-policy-binding payments-service-staging \
  --member="serviceAccount:github-deployer@payments-service-staging.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"
```

If needed, also grant read access for service config inspection:
```
gcloud projects add-iam-policy-binding payments-service-staging \
  --member="serviceAccount:github-deployer@payments-service-staging.iam.gserviceaccount.com" \
  --role="roles/viewer"
```

### 2. Create a Workload Identity Pool and Provider

Follow Google Cloud's Workload Identity Federation setup for GitHub Actions.

At the end of setup, you will have:
- a Workload Identity Provider resource name
- a deployer service account email that GitHub Actions can impersonate

#### Required GitHub repository secrets

Add these repository secrets in GitHub:

- `GCP_WORKLOAD_IDENTITY_PROVIDER`
- `GCP_DEPLOY_SERVICE_ACCOUNT`
- `GCP_RUNTIME_SERVICE_ACCOUNT`

#### Expected values

`GCP_WORKLOAD_IDENTITY_PROVIDER` should look like:
```
projects/123456789/locations/global/workloadIdentityPools/github/providers/github-provider
```

`GCP_DEPLOY_SERVICE_ACCOUNT` should look like:
```
github-deployer@payments-service-staging.iam.gserviceaccount.com
```

`GCP_RUNTIME_SERVICE_ACCOUNT` should look like:
```
payments-service@payments-service-staging.iam.gserviceaccount.com
```

#### Release version behavior

If the published GitHub Release tag is: `v1.0.0`


then the workflow will:
- push image tag `v1.0.0`
- deploy that image to Cloud Run
- set `APP_VERSION=v1.0.0`

This makes the version visible in:
- Artifact Registry
- Cloud Run deployment config
- /healthz
- structured logs
