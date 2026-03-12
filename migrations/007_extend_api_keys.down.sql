DROP INDEX IF EXISTS idx_api_keys_owner;
ALTER TABLE api_keys DROP COLUMN IF EXISTS updated_at;
ALTER TABLE api_keys DROP COLUMN IF EXISTS last_used_at;
ALTER TABLE api_keys DROP COLUMN IF EXISTS owner_id;
ALTER TABLE api_keys DROP COLUMN IF EXISTS key_prefix;
