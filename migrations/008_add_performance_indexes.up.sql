CREATE INDEX IF NOT EXISTS idx_documents_owner_created
    ON documents(owner_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_documents_content_type
    ON documents(content_type)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_documents_attachment_type
    ON documents(attachment_type_id, owner_class_name)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_sync_log_action_status
    ON sync_log(action, status, synced_at DESC);
