CREATE TABLE documents (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    minio_bucket        VARCHAR(63) NOT NULL,
    minio_key           VARCHAR(1024) NOT NULL,
    filename            VARCHAR(255) NOT NULL,
    content_type        VARCHAR(250),
    file_size           BIGINT NOT NULL,
    sha256_hash         CHAR(64) NOT NULL,
    owner_id            VARCHAR(32) NOT NULL,
    owner_class_library VARCHAR(512) NOT NULL,
    owner_class_name    VARCHAR(60) NOT NULL,
    attachment_type_id  INT,
    attachment_type     VARCHAR(255),
    is_external         BOOLEAN DEFAULT FALSE,
    legacy_file_id      VARCHAR(32),
    current_version     INT NOT NULL DEFAULT 1,
    created_by          VARCHAR(255),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    UNIQUE(minio_bucket, minio_key)
);

CREATE INDEX idx_documents_owner ON documents(owner_id, owner_class_library, owner_class_name);
CREATE INDEX idx_documents_hash ON documents(sha256_hash);
CREATE INDEX idx_documents_deleted ON documents(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_legacy ON documents(legacy_file_id, owner_id);
