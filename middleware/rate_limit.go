package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
)

type RateLimit interface {
	Limit(next http.HandlerFunc, limit int) http.HandlerFunc
	LimitAndRecordHit(next http.HandlerFunc, limit int, expiry time.Duration) http.HandlerFunc
	RecordHit(r *http.Request, expiry time.Duration)
}

type rateLimit struct {
	service services.RateLimitService
}

func NewRateLimit(service services.RateLimitService) RateLimit {
	return &rateLimit{
		service,
	}
}

// Limit checks if the request exceeds the rate limit.
func (m *rateLimit) Limit(next http.HandlerFunc, limit int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl := &types.RateLimit{
			IPAddress: getClientIP(r),
			Path:      r.URL.Path,
		}
		if err := m.service.GetHitCount(r.Context(), rl); err != nil {
			slog.Error("Error checking rate limit", "error", err)
			next(w, r)
			return
		}
		if rl.HitCount >= limit {
			// w.Header().Set("Retry-After", "3600") // Dynamically set based on your rate limit window
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next(w, r)
	})
}

// LimitAndRecordHit checks the rate limit and records a hit if under the limit.
func (m *rateLimit) LimitAndRecordHit(next http.HandlerFunc, limit int, expiry time.Duration) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl := &types.RateLimit{
			IPAddress: getClientIP(r),
			Path:      r.URL.Path,
			Limit:     limit,
			ExpiresAt: time.Now().UTC().Add(expiry),
		}
		if err := m.service.GetHitCount(r.Context(), rl); err != nil {
			slog.Error("Error checking rate limit", "error", err)
			next(w, r)
			return
		}
		if rl.HitCount >= limit {
			// w.Header().Set("Retry-After", "3600") // Dynamically set based on your rate limit window
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		if err := m.service.RecordHit(r.Context(), rl); err != nil {
			slog.Error("Error recording hit", "error", err)
		}
		next(w, r)
	})
}

// RecordHit logs a hit for the given request and expiry duration.
func (m *rateLimit) RecordHit(r *http.Request, expiry time.Duration) {
	rl := &types.RateLimit{
		IPAddress: getClientIP(r),
		Path:      r.URL.Path,
		ExpiresAt: time.Now().UTC().Add(expiry),
	}
	if err := m.service.RecordHit(r.Context(), rl); err != nil {
		slog.Error("Error recording hit", "error", err)
	}
}

// getClientIP extracts the client's IP address from the request.
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header first (common with proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0]) // First IP is the client
	}

	// Check X-Real-IP next (used by some proxies)
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Fall back to remote address
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
