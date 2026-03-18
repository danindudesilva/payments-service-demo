
## Running local performance tests

Install `k6`:
```bash
brew install k6
```

then run the service locally with the fake provider. Recommended local test configuration:

```bash
export APP_ENV=development
export PORT=3000
export DATABASE_URL="postgres://payments_service:payments_service@localhost:5432/payments_service?sslmode=disable"
export PAYMENTS_PROVIDER=fake
go run ./cmd/api
```

The fake provider should be used for:
- baseline load tests
- idempotency stress tests
- soak tests
- spike tests

Use the real Stripe-backed deployment only for low-volume integration and regression testing.

Run the baseline performance test:
```bash
make perf-baseline
```

Run the idempotency concurrency test:
```bash
make perf-idempotency
```

You can also point the same tests at another environment, such as Cloud Run:
```
BASE_URL="https://your-service-url.run.app" make perf-baseline
BASE_URL="https://your-service-url.run.app" make perf-idempotency
```

### What these tests cover

baseline.js:
- create payment attempt
- get payment attempt
- reconcile payment attempt

idempotency.js:
- concurrent requests using the same Idempotency-Key
- validates high-contention idempotent create behavior

### Recommended usage

Use `PAYMENTS_PROVIDER=fake` for load testing.

Do not use the real Stripe-backed environment for heavy load testing. Use the Stripe-backed environment only for low-volume integration and end-to-end correctness checks.

### Interpreting results

Good first signals to watch:

- low HTTP failure rate
- stable p95 latency
- correct behavior under concurrent idempotent create requests

If error rates rise under idempotency load, inspect:

- database contention
- unique constraint handling
- Cloud Run concurrency settings
- request timeouts
