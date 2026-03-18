import { sleep } from "k6";
import { SLEEP_SECONDS } from "./consts.js";

import {
  randomSuffix,
  createAttempt,
  getAttempt,
  reconcileAttempt,
  parseJSON,
} from "./helpers.js";

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

export default function () {
  const orderId = `order_perf_${randomSuffix()}`;
  const idempotencyKey = `idem_perf_${randomSuffix()}`;

  const createRes = createAttempt(orderId, idempotencyKey);
  const body = parseJSON(createRes);

  if (!body || !body.id) {
    return;
  }

  getAttempt(body.id);
  reconcileAttempt(body.id);

  sleep(SLEEP_SECONDS);
}
