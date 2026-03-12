package postgres_test

import (
	"testing"

	"github.com/jsi/ibs-doc-engine/internal/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateAPIKeyForTest is a helper that calls the unexported generateAPIKey via
// a thin exported wrapper we test through the public surface.
// Since generateAPIKey is unexported we test it indirectly via exported behaviour
// (we can verify its contract by calling it via reflection or by testing the
// exported Create method). For now we expose a thin test-only shim through
// the package's test-visible helper.

// TestGenerateAPIKey verifies prefix format, hash length, and prefix alignment.
func TestGenerateAPIKey(t *testing.T) {
	plaintext, hashHex, prefix, err := postgres.ExportedGenerateAPIKey()
	require.NoError(t, err)

	// plaintext starts with "ibsde_" and is 6+64=70 chars ("ibsde_" + 32 bytes hex = 6+64)
	assert.Equal(t, "ibsde_", plaintext[:6], "plaintext should start with ibsde_")
	assert.Len(t, plaintext, 70, "plaintext should be 70 chars")

	// hash is sha256 hex = 64 chars
	assert.Len(t, hashHex, 64, "hash should be 64 hex chars")

	// prefix is the first 8 chars of plaintext
	assert.Equal(t, plaintext[:8], prefix, "prefix should be first 8 chars of plaintext")
	assert.Len(t, prefix, 8, "prefix should be 8 chars")
}

// TestGenerateAPIKey_Uniqueness checks that 100 generated keys are all distinct.
func TestGenerateAPIKey_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		plaintext, hashHex, _, err := postgres.ExportedGenerateAPIKey()
		require.NoError(t, err)
		assert.NotContains(t, seen, plaintext, "duplicate plaintext key detected")
		assert.NotContains(t, seen, hashHex, "duplicate hash detected")
		seen[plaintext] = struct{}{}
		seen[hashHex] = struct{}{}
	}
}
