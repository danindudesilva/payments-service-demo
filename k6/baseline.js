import http from "k6/http";
import { check, sleep } from "k6";

const BASE_URL = __ENV.APP_URL || "http://localhost:3000";

export const options = {
  scenarios: {
    baseline_api: {
      executor: "ramping-vus",
      startVUs: 1,
      stages: [
        { duration: "30s", target: 5 },
        { duration: "1m", target: 10 },
        { duration: "30s", target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<1000"],
  },
};

function randomSuffix() {
  return Math.random().toString(36).slice(2, 10);
}

function createAttempt() {
  const orderId = `order_perf_${randomSuffix()}`;
  const idempotencyKey = `idem_perf_${randomSuffix()}`;

  const payload = JSON.stringify({
    order_id: orderId,
    amount: 3456,
    currency: "gbp",
    return_url: "https://example.com/return",
    description: "k6 baseline test payment",
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

function parseJSON(res) {
  try {
    return res.json();
  } catch (_) {
    return null;
  }
}

function getAttempt(attemptId) {
  const res = http.get(`${BASE_URL}/payment-attempts/${attemptId}`);

  check(res, {
    "get attempt status is 200": (r) => r.status === 200,
  });

  return res;
}

function reconcileAttempt(attemptId) {
  const res = http.post(
    `${BASE_URL}/payment-attempts/${attemptId}/reconcile`,
    null
  );

  check(res, {
    "reconcile attempt status is 200": (r) => r.status === 200,
  });

  return res;
}

export default function () {
  const createRes = createAttempt();
  const body = parseJSON(createRes);

  if (!body || !body.id) {
    return;
  }

  getAttempt(body.id);
  reconcileAttempt(body.id);

  sleep(1);
}
