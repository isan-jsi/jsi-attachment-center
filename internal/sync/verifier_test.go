package sync_test

import (
	"testing"

	syncsvc "github.com/jsi/ibs-doc-engine/internal/sync"
	"github.com/stretchr/testify/assert"
)

func TestVerify_HashMatch(t *testing.T) {
	err := syncsvc.Verify("abc123", 1024, "abc123", 1024)
	assert.NoError(t, err)
}

func TestVerify_HashMismatch(t *testing.T) {
	err := syncsvc.Verify("abc123", 1024, "xyz789", 1024)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash mismatch")
}

func TestVerify_SizeMismatch(t *testing.T) {
	err := syncsvc.Verify("abc123", 1024, "abc123", 2048)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "size mismatch")
}
