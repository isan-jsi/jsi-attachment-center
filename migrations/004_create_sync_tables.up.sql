CREATE TABLE sync_checkpoints (
    table_name          VARCHAR(100) PRIMARY KEY,
    last_dc_check       BYTEA,
    last_sync_at        TIMESTAMPTZ NOT NULL,
    records_processed   BIGINT DEFAULT 0,
    status              VARCHAR(20) DEFAULT 'active'
);

CREATE TABLE sync_log (
    id              UUID DEFAULT gen_random_uuid(),
    document_id     UUID,
    legacy_owner_id VARCHAR(32),
    legacy_file_id  VARCHAR(32),
    action          VARCHAR(20) NOT NULL CHECK (action IN ('create', 'update', 'delete', 'error')),
    status          VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'retrying', 'dlq')),
    error_message   TEXT,
    duration_ms     INT,
    synced_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, synced_at)
) PARTITION BY RANGE (synced_at);

-- Create initial partition for current month
CREATE TABLE sync_log_default PARTITION OF sync_log DEFAULT;

CREATE INDEX idx_sync_log_status ON sync_log(status, synced_at);

CREATE TABLE sync_dlq (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    legacy_owner_id VARCHAR(32) NOT NULL,
    legacy_file_id  VARCHAR(32) NOT NULL,
    table_name      VARCHAR(100) NOT NULL,
    retry_count     INT NOT NULL DEFAULT 0,
    max_retries     INT NOT NULL DEFAULT 5,
    next_retry_at   TIMESTAMPTZ,
    last_error      TEXT,
    payload_json    JSONB NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'retrying', 'exhausted', 'resolved')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(legacy_owner_id, legacy_file_id, table_name)
);

CREATE INDEX idx_sync_dlq_next_retry ON sync_dlq(next_retry_at) WHERE status IN ('pending', 'retrying');
