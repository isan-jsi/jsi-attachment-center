CREATE TABLE access_control (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     UUID REFERENCES documents(id),
    folder_id       UUID REFERENCES folders(id),
    principal_id    VARCHAR(255) NOT NULL,
    principal_type  VARCHAR(20) NOT NULL CHECK (principal_type IN ('user', 'role')),
    permission      VARCHAR(20) NOT NULL CHECK (permission IN ('read', 'write', 'admin')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (document_id IS NOT NULL OR folder_id IS NOT NULL)
);

CREATE INDEX idx_access_principal ON access_control(principal_id, principal_type);
CREATE INDEX idx_access_document ON access_control(document_id);
CREATE INDEX idx_access_folder ON access_control(folder_id);
