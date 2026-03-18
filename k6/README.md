
## Running local performance tests

Install `k6`:
```bash
brew install k6
```

then run the service locally with the fake provider. Recommended local test configuration:

```bash
export APP_ENV=development
export APP_URL=3000
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
APP_URL="https://your-service-url.run.app" make perf-baseline
APP_URL="https://your-service-url.run.app" make perf-idempotency
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

### Verifying idempotency after the k6 concurrency test

After running the idempotency test, verify that all concurrent requests with the same `Idempotency-Key` converged to a single payment attempt.

Example query:

```sql
SELECT id, order_id, idempotency_key, provider_name, provider_payment_id, status
FROM payment_attempts
WHERE idempotency_key = 'idem_shared_key';
```
Expected result:
- exactly one row is returned

This confirms that:
- concurrent duplicate create requests were handled idempotently
- the database uniqueness constraints were effective
- the service replayed the existing attempt instead of creating duplicates
