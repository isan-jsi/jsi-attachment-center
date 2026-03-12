package minio

import "testing"

func BenchmarkComputeSHA256(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
	}
	for _, s := range sizes {
		data := make([]byte, s.size)
		b.Run(s.name, func(b *testing.B) {
			b.SetBytes(int64(s.size))
			for i := 0; i < b.N; i++ {
				ComputeSHA256(data)
			}
		})
	}
}

func BenchmarkDetectContentType(b *testing.B) {
	data := []byte("%PDF-1.4 fake pdf content here for testing purposes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectContentType(data, "report.pdf")
	}
}
