package server

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const maxRateLimitClients = 10_000

type rateLimiter struct {
	mu      sync.Mutex
	limit   rate.Limit
	burst   int
	clients map[string]rateLimitEntry
}

type rateLimitEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		limit:   rate.Limit(float64(limit) / window.Seconds()),
		burst:   limit,
		clients: make(map[string]rateLimitEntry),
	}
}

func (l *rateLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.clients[key]
	if !ok {
		if len(l.clients) >= maxRateLimitClients {
			l.evictOldest()
		}
		entry = rateLimitEntry{limiter: rate.NewLimiter(l.limit, l.burst)}
	}
	entry.lastSeen = now

	reservation := entry.limiter.ReserveN(now, 1)
	if !reservation.OK() {
		l.clients[key] = entry
		return false, 0
	}

	if wait := reservation.DelayFrom(now); wait > 0 {
		reservation.CancelAt(now)
		l.clients[key] = entry
		return false, wait
	}

	l.clients[key] = entry
	return true, 0
}

func (l *rateLimiter) evictOldest() {
	var oldestKey string
	var oldest time.Time
	for key, entry := range l.clients {
		if oldestKey == "" || entry.lastSeen.Before(oldest) {
			oldestKey = key
			oldest = entry.lastSeen
		}
	}
	if oldestKey != "" {
		delete(l.clients, oldestKey)
	}
}

func (app *app) rateLimitOIDC(next http.Handler) http.Handler {
	if app.oidcRateLimiter == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path + "\x00" + requestClientAddress(r)
		allowed, retryAfter := app.oidcRateLimiter.allow(key, time.Now())
		if !allowed {
			w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds(retryAfter), 10))
			app.errorResponse(w, r, http.StatusTooManyRequests, "rate_limited", "Too many authentication attempts. Try again later")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func requestClientAddress(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return "unknown"
}

func retryAfterSeconds(wait time.Duration) int64 {
	seconds := int64(wait / time.Second)
	if wait%time.Second != 0 {
		seconds++
	}
	return max(int64(1), seconds)
}
