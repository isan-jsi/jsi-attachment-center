import { api, buildQueryString } from "./client";
import type { ApiEnvelope } from "@/types/api";
import type { Document } from "@/types/document";

const BASE = "/api/v1/search";

export interface SearchParams {
  q: string;
  folder_id?: number;
  owner_class_library?: string;
  owner_class_name?: string;
  document_type?: string;
  page?: number;
  page_size?: number;
}

export const searchApi = {
  search(params: SearchParams): Promise<ApiEnvelope<Document[]>> {
    return api.get<Document[]>(`${BASE}${buildQueryString(params)}`);
  },
};
