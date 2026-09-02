package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowsLimitThenBlocks(t *testing.T) {
	limiter := newRateLimiter(2, time.Minute)
	now := time.Now()

	for range 2 {
		allowed, retryAfter := limiter.allow("client", now)
		if !allowed {
			t.Fatal("request was blocked before the limit was reached")
		}
		if retryAfter != 0 {
			t.Fatalf("retryAfter = %s, want zero for an allowed request", retryAfter)
		}
	}

	allowed, retryAfter := limiter.allow("client", now)
	if allowed {
		t.Fatal("request was allowed after the limit was reached")
	}
	if retryAfter <= 0 || retryAfter > time.Minute {
		t.Fatalf("retryAfter = %s, want a duration between 0 and 1m", retryAfter)
	}
}

func TestRateLimiterRefillsTokensBeforeTheWindowEnds(t *testing.T) {
	limiter := newRateLimiter(2, time.Minute)
	now := time.Now()

	for range 2 {
		allowed, _ := limiter.allow("client", now)
		if !allowed {
			t.Fatal("request was blocked before the burst was exhausted")
		}
	}

	allowed, _ := limiter.allow("client", now.Add(30*time.Second))
	if !allowed {
		t.Fatal("request was blocked after a token should have refilled")
	}
}

func TestRateLimitMiddlewareUsesValidatedRemoteAddress(t *testing.T) {
	limiter := newRateLimiter(1, time.Minute)
	called := 0
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called++ })
	app := testApp()
	app.oidcRateLimiter = limiter
	handler := app.rateLimitOIDC(next)

	first := httptest.NewRequest(http.MethodGet, "/v1/auth/google/start", nil)
	first.RemoteAddr = "192.0.2.1:1234"
	first.Header.Set("X-Forwarded-For", "198.51.100.1")
	firstResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, first)

	second := httptest.NewRequest(http.MethodGet, "/v1/auth/google/start", nil)
	second.RemoteAddr = "192.0.2.1:1234"
	second.Header.Set("X-Forwarded-For", "203.0.113.1")
	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(secondResponse, second)

	if called != 1 {
		t.Fatalf("next handler called %d times, want 1", called)
	}
	if secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf("status code = %d, want %d", secondResponse.Code, http.StatusTooManyRequests)
	}
	if secondResponse.Header().Get("Retry-After") == "" {
		t.Fatal("Retry-After header is empty")
	}
}

func TestRateLimiterResetsAfterWindow(t *testing.T) {
	limiter := newRateLimiter(1, time.Minute)
	now := time.Now()

	limiter.allow("client", now)
	allowed, _ := limiter.allow("client", now.Add(time.Minute))
	if !allowed {
		t.Fatal("request was not allowed after the window reset")
	}
}
