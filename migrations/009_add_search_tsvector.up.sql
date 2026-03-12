-- Add tsvector column for full-text search
ALTER TABLE documents ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Populate from existing data
UPDATE documents SET search_vector =
    setweight(to_tsvector('english', coalesce(filename, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(attachment_type, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(owner_class_name, '')), 'C');

-- GIN index for fast search
CREATE INDEX IF NOT EXISTS idx_documents_search ON documents USING GIN(search_vector);

-- Trigger to auto-update tsvector on INSERT/UPDATE
CREATE OR REPLACE FUNCTION documents_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', coalesce(NEW.filename, '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.attachment_type, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.owner_class_name, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_documents_search_vector
    BEFORE INSERT OR UPDATE OF filename, attachment_type, owner_class_name
    ON documents
    FOR EACH ROW
    EXECUTE FUNCTION documents_search_vector_update();
