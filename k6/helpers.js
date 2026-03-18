import http from "k6/http";
import { check } from "k6";

import { AMOUNT, BASE_URL } from "./consts.js";

export function createAttempt(orderId, idempotencyKey) {
  const payload = JSON.stringify({
    order_id: orderId,
    amount: AMOUNT,
    currency: "gbp",
    return_url: "https://example.com/return",
    description: "k6 test payment",
  });

  const res = http.post(`${BASE_URL}/payment-attempts`, payload, {
    headers: {
      "Content-Type": "application/json",
      "Idempotency-Key": idempotencyKey,
    },
  });

  check(res, {
    "create attempt status is 200 or 201": (r) =>
      r.status === 200 || r.status === 201,
  });

  return res;
}

export function getAttempt(attemptId) {
  const res = http.get(`${BASE_URL}/payment-attempts/${attemptId}`);

  check(res, {
    "get attempt status is 200": (r) => r.status === 200,
  });

  return res;
}

export function reconcileAttempt(attemptId) {
  const res = http.post(
    `${BASE_URL}/payment-attempts/${attemptId}/reconcile`,
    null
  );

  check(res, {
    "reconcile attempt status is 200": (r) => r.status === 200,
  });

  return res;
}

export function randomSuffix() {
  return Math.random().toString(36).slice(2, 10);
}

export function parseJSON(res) {
  try {
    return res.json();
  } catch (_) {
    return null;
  }
}
