import http from "k6/http";
import { check } from "k6";

const BASE_URL = __ENV.APP_URL || "http://localhost:3000";
const SHARED_ORDER_ID = __ENV.ORDER_ID || "order_idem_shared";
const SHARED_IDEMPOTENCY_KEY = __ENV.IDEMPOTENCY_KEY || "idem_shared_key";
const AMOUNT = Number(__ENV.AMOUNT || "3456");

export const options = {
  scenarios: {
    idempotency_race: {
      executor: "constant-vus",
      vus: 20,
      duration: "20s",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<1000"],
  },
};

export default function () {
  const payload = JSON.stringify({
    order_id: SHARED_ORDER_ID,
    amount: AMOUNT,
    currency: "gbp",
    return_url: "https://example.com/return",
    description: "k6 idempotency test payment",
  });

  const res = http.post(`${BASE_URL}/payment-attempts`, payload, {
    headers: {
      "Content-Type": "application/json",
      "Idempotency-Key": SHARED_IDEMPOTENCY_KEY,
    },
  });

  check(res, {
    "idempotent create returns success": (r) =>
      r.status === 200 || r.status === 201,
  });
}
