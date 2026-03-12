import type { ListParams } from "./api";

export interface Document {
  id: number;
  folder_id: number;
  owner_class_library: string;
  owner_class_name: string;
  document_type: string;
  title: string;
  file_name: string;
  file_extension: string;
  file_size: number;
  mime_type: string;
  storage_path: string;
  checksum: string;
  version: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  folder_name?: string;
}

export interface DocumentCreate {
  folder_id: number;
  owner_class_library: string;
  owner_class_name: string;
  document_type: string;
  title: string;
  file: File;
}

export interface DocumentUpdate {
  title?: string;
  folder_id?: number;
  is_active?: boolean;
}

export interface DocumentListParams extends ListParams {
  folder_id?: number;
  owner_class_library?: string;
  owner_class_name?: string;
  document_type?: string;
  is_active?: boolean;
  q?: string;
}
