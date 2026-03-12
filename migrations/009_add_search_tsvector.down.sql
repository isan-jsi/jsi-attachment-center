DROP TRIGGER IF EXISTS trg_documents_search_vector ON documents;
DROP FUNCTION IF EXISTS documents_search_vector_update();
DROP INDEX IF EXISTS idx_documents_search;
ALTER TABLE documents DROP COLUMN IF EXISTS search_vector;
