import {
  useQuery,
  useMutation,
  useQueryClient,
  keepPreviousData,
} from "@tanstack/react-query";
import { documentsApi } from "@/api/documents";
import type {
  DocumentCreate,
  DocumentUpdate,
  DocumentListParams,
} from "@/types/document";
import { folderKeys } from "./use-folders";

export const documentKeys = {
  all: ["documents"] as const,
  list: (params: DocumentListParams) => ["documents", "list", params] as const,
  detail: (id: number) => ["documents", id] as const,
};

export function useDocuments(params: DocumentListParams = {}) {
  return useQuery({
    queryKey: documentKeys.list(params),
    queryFn: () => documentsApi.list(params),
    placeholderData: keepPreviousData,
  });
}

export function useDocument(id: number) {
  return useQuery({
    queryKey: documentKeys.detail(id),
    queryFn: () => documentsApi.get(id).then((r) => r.data),
    enabled: !!id,
  });
}

export function useCreateDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: DocumentCreate) =>
      documentsApi.create(data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: documentKeys.all });
      qc.invalidateQueries({ queryKey: folderKeys.all });
    },
  });
}

export function useUpdateDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: DocumentUpdate }) =>
      documentsApi.update(id, data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: documentKeys.all });
    },
  });
}

export function useDeleteDocument() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => documentsApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: documentKeys.all });
      qc.invalidateQueries({ queryKey: folderKeys.all });
    },
  });
}

export function useDeleteDocuments() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (ids: number[]) => {
      await Promise.all(ids.map((id) => documentsApi.delete(id)));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: documentKeys.all });
      qc.invalidateQueries({ queryKey: folderKeys.all });
    },
  });
}

export function getDownloadUrl(id: number): string {
  return documentsApi.download(id);
}
