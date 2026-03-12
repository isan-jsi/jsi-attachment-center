import http from "k6/http";
import { check, sleep } from "k6";
import { Rate } from "k6/metrics";

const errorRate = new Rate("errors");

export const options = {
  stages: [
    { duration: "30s", target: 10 },
    { duration: "1m",  target: 50 },
    { duration: "30s", target: 100 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"],
    errors: ["rate<0.05"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const API_KEY = __ENV.API_KEY || "test-api-key";
const headers = { "X-API-Key": API_KEY };

export default function () {
  const listRes = http.get(`${BASE_URL}/api/v1/documents?page=1&per_page=20`, { headers });
  check(listRes, { "list: status 200": (r) => r.status === 200 });
  errorRate.add(listRes.status !== 200);
  sleep(0.5);

  const healthRes = http.get(`${BASE_URL}/health`);
  check(healthRes, { "health: status 200": (r) => r.status === 200 });
  sleep(0.5);
}
