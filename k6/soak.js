import { sleep } from "k6";
import { SLEEP_SECONDS } from "./consts.js";

import {
  randomSuffix,
  createAttempt,
  getAttempt,
  reconcileAttempt,
  parseJSON,
} from "./helpers.js";

const VUS = Number(__ENV.VUS || "10");
const DURATION = __ENV.DURATION || "30m";

export const options = {
  scenarios: {
    soak_api: {
      executor: "constant-vus",
      vus: VUS,
      duration: DURATION,
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<1000"],
  },
};

export default function () {
  const orderId = `order_soak_${randomSuffix()}`;
  const idempotencyKey = `idem_soak_${randomSuffix()}`;

  const createRes = createAttempt(orderId, idempotencyKey);
  const body = parseJSON(createRes);

  if (body && body.id) {
    getAttempt(body.id);
    reconcileAttempt(body.id);
  }

  sleep(SLEEP_SECONDS);
}
