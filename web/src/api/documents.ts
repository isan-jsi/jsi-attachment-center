import { api, buildQueryString } from "./client";
import type { ApiEnvelope } from "@/types/api";
import type {
  Document,
  DocumentCreate,
  DocumentUpdate,
  DocumentListParams,
} from "@/types/document";

const BASE = "/api/v1/documents";

export const documentsApi = {
  list(params: DocumentListParams = {}): Promise<ApiEnvelope<Document[]>> {
    return api.get<Document[]>(`${BASE}${buildQueryString(params)}`);
  },

  get(id: number): Promise<ApiEnvelope<Document>> {
    return api.get<Document>(`${BASE}/${id}`);
  },

  create(data: DocumentCreate): Promise<ApiEnvelope<Document>> {
    const formData = new FormData();
    formData.append("folder_id", String(data.folder_id));
    formData.append("owner_class_library", data.owner_class_library);
    formData.append("owner_class_name", data.owner_class_name);
    formData.append("document_type", data.document_type);
    formData.append("title", data.title);
    formData.append("file", data.file);
    return api.post<Document>(BASE, formData);
  },

  update(id: number, data: DocumentUpdate): Promise<ApiEnvelope<Document>> {
    return api.patch<Document>(`${BASE}/${id}`, data);
  },

  delete(id: number): Promise<ApiEnvelope<void>> {
    return api.delete<void>(`${BASE}/${id}`);
  },

  download(id: number): string {
    return `${BASE}/${id}/download`;
  },
};
