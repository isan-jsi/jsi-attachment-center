export interface Folder {
  id: number;
  parent_id: number | null;
  name: string;
  path: string;
  description: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  children?: Folder[];
  document_count?: number;
}

export interface FolderCreate {
  parent_id?: number | null;
  name: string;
  description?: string;
}

export interface FolderUpdate {
  name?: string;
  description?: string;
  is_active?: boolean;
}
