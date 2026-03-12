package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// okHandler is a simple handler that always returns 200 OK.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// TestRateLimiter_AllowsBurst verifies that exactly `burst` requests are allowed,
// and the (burst+1)-th request is rejected with HTTP 429.
func TestRateLimiter_AllowsBurst(t *testing.T) {
	const burst = 5
	rl := NewRateLimiter(0.0001, burst) // near-zero RPS so tokens don't refill during test
	handler := rl.Middleware()(okHandler)

	for i := 0; i < burst; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	// (burst+1)-th request should be rate-limited.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on burst+1 request, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429 response")
	}
}

// TestRateLimiter_DifferentKeysIndependent ensures that two different IP addresses
// have independent token buckets.
func TestRateLimiter_DifferentKeysIndependent(t *testing.T) {
	const burst = 2
	rl := NewRateLimiter(0.0001, burst)
	handler := rl.Middleware()(okHandler)

	// Exhaust quota for IP A.
	for i := 0; i < burst; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:9999"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("IP-A request %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	// IP A should now be limited.
	reqA := httptest.NewRequest(http.MethodGet, "/", nil)
	reqA.RemoteAddr = "192.168.1.1:9999"
	rrA := httptest.NewRecorder()
	handler.ServeHTTP(rrA, reqA)
	if rrA.Code != http.StatusTooManyRequests {
		t.Fatalf("IP-A: expected 429, got %d", rrA.Code)
	}

	// IP B should still have a full bucket.
	reqB := httptest.NewRequest(http.MethodGet, "/", nil)
	reqB.RemoteAddr = "192.168.1.2:9999"
	rrB := httptest.NewRecorder()
	handler.ServeHTTP(rrB, reqB)
	if rrB.Code != http.StatusOK {
		t.Fatalf("IP-B: expected 200, got %d", rrB.Code)
	}
}

// TestExtractKey_APIKeyHeader verifies that X-API-Key header takes precedence.
func TestExtractKey_APIKeyHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "my-secret-key")
	req.RemoteAddr = "10.0.0.1:1234"

	key := extractKey(req)
	expected := "apikey:my-secret-key"
	if key != expected {
		t.Errorf("extractKey: got %q, want %q", key, expected)
	}
}

// TestExtractKey_XForwardedFor verifies that X-Forwarded-For is used when no API key is present.
func TestExtractKey_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.42")
	req.RemoteAddr = "10.0.0.1:1234"

	key := extractKey(req)
	expected := "ip:203.0.113.42"
	if key != expected {
		t.Errorf("extractKey: got %q, want %q", key, expected)
	}
}

// TestExtractKey_RemoteAddr verifies that RemoteAddr is used as a fallback.
func TestExtractKey_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.16.0.5:8888"

	key := extractKey(req)
	expected := "ip:172.16.0.5:8888"
	if key != expected {
		t.Errorf("extractKey: got %q, want %q", key, expected)
	}
}
