CREATE TABLE folders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(255) NOT NULL,
    parent_id           UUID REFERENCES folders(id),
    path                TEXT NOT NULL,
    owner_class_library VARCHAR(512),
    owner_class_name    VARCHAR(60),
    owner_id            VARCHAR(32),
    is_auto_generated   BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_folders_parent ON folders(parent_id);
CREATE INDEX idx_folders_path ON folders USING btree (path text_pattern_ops);
CREATE UNIQUE INDEX idx_folders_path_unique ON folders(path);

CREATE TABLE folder_documents (
    folder_id   UUID NOT NULL REFERENCES folders(id),
    document_id UUID NOT NULL REFERENCES documents(id),
    sort_order  INT DEFAULT 0,
    PRIMARY KEY (folder_id, document_id)
);

CREATE TABLE document_versions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     UUID NOT NULL REFERENCES documents(id),
    version_number  INT NOT NULL,
    minio_key       VARCHAR(1024) NOT NULL,
    content_type    VARCHAR(250),
    sha256_hash     CHAR(64) NOT NULL,
    file_size       BIGINT NOT NULL,
    created_by      VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(document_id, version_number)
);
