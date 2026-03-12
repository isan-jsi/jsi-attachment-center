package domain

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID                uuid.UUID
	MinioBucket       string
	MinioKey          string
	Filename          string
	ContentType       string
	FileSize          int64
	SHA256Hash        string
	OwnerID           string
	OwnerClassLibrary string
	OwnerClassName    string
	AttachmentTypeID  int
	AttachmentType    string
	IsExternal        bool
	LegacyFileID      string
	CurrentVersion    int
	CreatedBy         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

// LegacyAttachment represents a row from IBSDocAttachments in SQL Server.
type LegacyAttachment struct {
	OwnerID             string
	FileID              string
	DocAttachmentTypeID int
	DocAttachmentType   string
	FileName            string
	ContentType         string
	FileContent         []byte // varbinary(MAX)
	FileSize            int64
	IsExternal          bool
	IsPostLockUpdate    bool
	CreatedBy           string
	CreatedOn           time.Time
	LastUpdateBy        string
	LastUpdateOn        time.Time
	DCCheck             []byte // SQL Server timestamp (8 bytes)
}

// LegacyOwner represents a row from IBSDocAttachmentOwners.
type LegacyOwner struct {
	OwnerClassLibrary     string
	OwnerClassName        string
	OwnerClassDescription string
}

// LegacyAttachmentType represents a row from IBSDocAttachmentTypes.
type LegacyAttachmentType struct {
	OwnerClassLibrary    string
	OwnerClassName       string
	DocAttachmentTypeID  int
	DocAttachmentType    string
	IsMandatory          bool
	AllowedFileExtension string
	MaxFileSizeInBytes   int64
}
