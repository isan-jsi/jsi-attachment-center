import {
  useQuery,
  useMutation,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query";
import { syncApi } from "@/api/sync";
import type { ListParams } from "@/types/api";

export const syncKeys = {
  all: ["sync"] as const,
  status: ["sync", "status"] as const,
  logs: (params: ListParams) => ["sync", "logs", params] as const,
  dlq: (params: ListParams) => ["sync", "dlq", params] as const,
};

export function useSyncStatus() {
  return useQuery({
    queryKey: syncKeys.status,
    queryFn: () => syncApi.status().then((r) => r.data),
    refetchInterval: 10_000,
  });
}

export function useSyncLogs(params: ListParams = {}) {
  return useQuery({
    queryKey: syncKeys.logs(params),
    queryFn: () => syncApi.logs(params),
    placeholderData: keepPreviousData,
  });
}

export function useSyncDlq(params: ListParams = {}) {
  return useQuery({
    queryKey: syncKeys.dlq(params),
    queryFn: () => syncApi.dlq(params),
    placeholderData: keepPreviousData,
  });
}

export function useTriggerSync() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => syncApi.trigger(),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: syncKeys.all });
    },
  });
}

export function useRetryDlq() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => syncApi.retryDlq(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: syncKeys.all });
    },
  });
}

export function useRetryAllDlq() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => syncApi.retryAllDlq(),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: syncKeys.all });
    },
  });
}
