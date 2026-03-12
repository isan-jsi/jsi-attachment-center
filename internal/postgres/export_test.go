package postgres

// ExportedGenerateAPIKey exposes the unexported generateAPIKey function for testing.
func ExportedGenerateAPIKey() (string, string, string, error) {
	return generateAPIKey()
}
