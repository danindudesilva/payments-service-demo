# Deployed regression checklist

Run these checks after deploying the service to Cloud Run.

## 1. Health check

Verify the service responds:

```bash
curl -i "https://<service-url>/healthz"
```
Expected:

- `200 OK`
- JSON body with `"status":"ok"`

## 2. Demo page

Open:
```text
https://<service-url>/demo
```
Expected:

- demo page renders
- Stripe Elements loads when configured

Verify in the demo page that:

- a fresh page load pre-populates a new random `order_id`
- amount defaults to a human-readable value such as `34.56`
- creating a payment sends the correct minor-unit amount to the backend
- resetting the demo generates a fresh default `order_id`

## 3. Create payment attempt

Use Postman or curl to create a payment attempt.

Verify:

- 201 Created
- attempt ID returned
- provider payment ID created after Stripe interaction
- idempotency header is honored

## 4. Full Stripe test-mode payment flow

- From the demo page:
- create a payment attempt
- confirm the payment with Stripe.js
- complete the payment using Stripe test mode
- return to the app

Verify:

- local payment attempt reaches succeeded
- provider payment ID is stored
- Stripe webhook events are accepted

## 5. Webhook verification

Use Stripe CLI or the Stripe dashboard webhook delivery logs.

Verify:
- payment_intent.succeeded returns 200
- duplicate event delivery returns 200
- duplicate delivery does not reprocess the event

## 6. Database verification

Check the database:

- payment_attempts contains the created attempt
- processed_webhook_events contains the Stripe event IDs

## 7. Idempotency verification

Replay the same create request with the same Idempotency-Key.

Verify:
- no duplicate payment attempt is created
- existing attempt is returned

## 8. Negative checks

Verify:

- missing required config causes startup failure
- invalid webhook signature returns 400
- unknown/unhandled Stripe event types return 200
