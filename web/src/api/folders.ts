import { api, buildQueryString } from "./client";
import type { ApiEnvelope, ListParams } from "@/types/api";
import type { Folder, FolderCreate, FolderUpdate } from "@/types/folder";

const BASE = "/api/v1/folders";

export const foldersApi = {
  list(params: ListParams = {}): Promise<ApiEnvelope<Folder[]>> {
    return api.get<Folder[]>(`${BASE}${buildQueryString(params)}`);
  },

  tree(): Promise<ApiEnvelope<Folder[]>> {
    return api.get<Folder[]>(`${BASE}/tree`);
  },

  get(id: number): Promise<ApiEnvelope<Folder>> {
    return api.get<Folder>(`${BASE}/${id}`);
  },

  create(data: FolderCreate): Promise<ApiEnvelope<Folder>> {
    return api.post<Folder>(BASE, data);
  },

  update(id: number, data: FolderUpdate): Promise<ApiEnvelope<Folder>> {
    return api.patch<Folder>(`${BASE}/${id}`, data);
  },

  delete(id: number): Promise<ApiEnvelope<void>> {
    return api.delete<void>(`${BASE}/${id}`);
  },
};
