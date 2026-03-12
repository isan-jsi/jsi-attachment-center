import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { foldersApi } from "@/api/folders";
import type { FolderCreate, FolderUpdate } from "@/types/folder";

export const folderKeys = {
  all: ["folders"] as const,
  tree: ["folders", "tree"] as const,
  detail: (id: number) => ["folders", id] as const,
};

export function useFolders() {
  return useQuery({
    queryKey: folderKeys.all,
    queryFn: () => foldersApi.list().then((r) => r.data),
  });
}

export function useFolderTree() {
  return useQuery({
    queryKey: folderKeys.tree,
    queryFn: () => foldersApi.tree().then((r) => r.data),
  });
}

export function useFolder(id: number) {
  return useQuery({
    queryKey: folderKeys.detail(id),
    queryFn: () => foldersApi.get(id).then((r) => r.data),
    enabled: !!id,
  });
}

export function useCreateFolder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: FolderCreate) =>
      foldersApi.create(data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: folderKeys.all });
      qc.invalidateQueries({ queryKey: folderKeys.tree });
    },
  });
}

export function useUpdateFolder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: FolderUpdate }) =>
      foldersApi.update(id, data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: folderKeys.all });
      qc.invalidateQueries({ queryKey: folderKeys.tree });
    },
  });
}

export function useDeleteFolder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => foldersApi.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: folderKeys.all });
      qc.invalidateQueries({ queryKey: folderKeys.tree });
    },
  });
}
