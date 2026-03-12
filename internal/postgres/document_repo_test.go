package postgres_test

import (
	"testing"

	"github.com/jsi/ibs-doc-engine/internal/postgres"
	"github.com/stretchr/testify/assert"
)

func TestDocumentRepo_Interface(t *testing.T) {
	repo := postgres.NewDocumentRepo(nil)
	assert.NotNil(t, repo)
}
