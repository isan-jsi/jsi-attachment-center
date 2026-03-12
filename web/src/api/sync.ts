import { api, buildQueryString } from "./client";
import type { ApiEnvelope, ListParams } from "@/types/api";
import type { SyncStatus, DLQEntry, SyncLogEntry } from "@/types/sync";

const BASE = "/api/v1/sync";

export const syncApi = {
  status(): Promise<ApiEnvelope<SyncStatus>> {
    return api.get<SyncStatus>(`${BASE}/status`);
  },

  trigger(): Promise<ApiEnvelope<void>> {
    return api.post<void>(`${BASE}/trigger`);
  },

  logs(params: ListParams = {}): Promise<ApiEnvelope<SyncLogEntry[]>> {
    return api.get<SyncLogEntry[]>(`${BASE}/logs${buildQueryString(params)}`);
  },

  dlq(params: ListParams = {}): Promise<ApiEnvelope<DLQEntry[]>> {
    return api.get<DLQEntry[]>(`${BASE}/dlq${buildQueryString(params)}`);
  },

  retryDlq(id: number): Promise<ApiEnvelope<void>> {
    return api.post<void>(`${BASE}/dlq/${id}/retry`);
  },

  retryAllDlq(): Promise<ApiEnvelope<void>> {
    return api.post<void>(`${BASE}/dlq/retry-all`);
  },
};
