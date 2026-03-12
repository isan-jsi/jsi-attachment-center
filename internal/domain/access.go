package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccessControl struct {
	ID            uuid.UUID
	DocumentID    *uuid.UUID
	FolderID      *uuid.UUID
	PrincipalID   string
	PrincipalType string // user, role
	Permission    string // read, write, admin
	CreatedAt     time.Time
}

// LegacyAccess represents a row from IBSDocAttachmentAccess.
type LegacyAccess struct {
	OwnerID   string
	FileID    string
	AccessID  string
	CreatedBy string
	CreatedOn time.Time
}
