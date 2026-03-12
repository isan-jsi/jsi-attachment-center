import http from "k6/http";
import { check, sleep } from "k6";
import { Rate, Trend } from "k6/metrics";

const uploadDuration = new Trend("upload_duration");
const errorRate = new Rate("errors");

export const options = {
  stages: [
    { duration: "30s", target: 5 },
    { duration: "1m",  target: 10 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    upload_duration: ["p(95)<2000"],
    errors: ["rate<0.10"],
  },
};

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const API_KEY = __ENV.API_KEY || "test-api-key";

export default function () {
  const fileData = open("/dev/urandom", "b").slice(0, 50 * 1024); // 50KB
  const fileName = `load-test-${Date.now()}.pdf`;
  const res = http.post(
    `${BASE_URL}/api/v1/documents`,
    { file: http.file(fileData, fileName, "application/pdf"), owner_id: "LOAD_TEST_OWNER" },
    { headers: { "X-API-Key": API_KEY } }
  );
  check(res, { "upload: status 201": (r) => r.status === 201 });
  uploadDuration.add(res.timings.duration);
  errorRate.add(res.status !== 201);
  sleep(2);
}
