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
