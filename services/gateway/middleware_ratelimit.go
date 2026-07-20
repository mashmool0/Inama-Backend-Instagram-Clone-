package main

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// rateLimitScript does an atomic INCR + (first-hit) EXPIRE so the counter and
// its TTL are set in one round trip — no race where a key lives forever because
// the process died between INCR and EXPIRE.
var rateLimitScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('EXPIRE', KEYS[1], 2)
end
return count
`)

// rateLimit is a fixed-window limiter: at most `limit` requests per client IP
// per one-second window. Simpler than a sliding window; a small boundary burst
// is acceptable here. If Redis is unavailable we fail OPEN (allow the request)
// rather than take the whole gateway down.
func rateLimit(rdb *redis.Client, limit int, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rdb == nil || limit <= 0 {
			next.ServeHTTP(w, r)
			return
		}
		ip := clientIP(r)
		key := "rl:" + ip + ":" + strconv.FormatInt(time.Now().Unix(), 10)

		count, err := rateLimitScript.Run(r.Context(), rdb, []string{key}, limit).Int()
		if err == nil && count > limit {
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP prefers the first X-Forwarded-For hop (we sit behind nginx in prod)
// and falls back to the direct socket address.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
