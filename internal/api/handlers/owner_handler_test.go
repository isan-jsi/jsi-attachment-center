package handlers_test

import (
	"testing"

	"github.com/jsi/ibs-doc-engine/internal/api/handlers"
	"github.com/stretchr/testify/assert"
)

func TestNewOwnerHandler_NotNil(t *testing.T) {
	h := handlers.NewOwnerHandler(nil)
	assert.NotNil(t, h)
}
