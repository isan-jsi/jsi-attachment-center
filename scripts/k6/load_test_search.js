import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

const searchDuration = new Trend("search_duration");
const errorRate = new Rate("errors");

export const options = {
  stages: [
    { duration: "30s", target: 20 },
    { duration: "2m",  target: 50 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    search_duration: ["p(95)<300"],
    errors: ["rate<0.05"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const API_KEY = __ENV.API_KEY || "test-api-key";
const headers = { "X-API-Key": API_KEY };
const searchTerms = ["invoice", "report", "contract", "attachment", "memo", "policy"];

export default function () {
  const term = searchTerms[Math.floor(Math.random() * searchTerms.length)];
  const res = http.get(`${BASE_URL}/api/v1/search?q=${term}&per_page=20`, { headers });
  check(res, { "search: status 200": (r) => r.status === 200 });
  searchDuration.add(res.timings.duration);
  errorRate.add(res.status !== 200);

  const prefix = term.substring(0, 3);
  const suggestRes = http.get(`${BASE_URL}/api/v1/search/suggest?q=${prefix}`, { headers });
  check(suggestRes, { "suggest: status 200": (r) => r.status === 200 });
  sleep(1);
}
