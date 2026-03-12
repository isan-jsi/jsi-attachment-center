package sync

import "fmt"

// Verify checks that uploaded content matches the expected hash and size.
func Verify(expectedHash string, expectedSize int64, actualHash string, actualSize int64) error {
	if expectedHash != actualHash {
		return fmt.Errorf("verification failed: hash mismatch (expected %s, got %s)", expectedHash, actualHash)
	}
	if expectedSize != actualSize {
		return fmt.Errorf("verification failed: size mismatch (expected %d, got %d)", expectedSize, actualSize)
	}
	return nil
}
