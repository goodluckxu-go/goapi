package middlewares

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/goodluckxu-go/goapi/v2"
)

const defaultRateLimitMessage = "too many requests"

// RateLimitConfig configures the token-bucket rate limit middleware.
type RateLimitConfig struct {
	// Limit is the number of requests allowed per Window.
	Limit int
	// Window is the time span used to refill Limit tokens.
	Window time.Duration
	// Burst is the maximum number of requests allowed at once.
	// If Burst is zero, Limit is used.
	Burst int
	// KeyFunc returns the bucket key for a request.
	// If KeyFunc is nil or returns an empty string, the direct remote IP is used.
	KeyFunc func(ctx *goapi.Context) string
	// Message is written in the 429 response body.
	Message string
	// CleanupInterval controls how often idle buckets are removed.
	// If zero, a conservative default derived from Window is used.
	CleanupInterval time.Duration
	// MaxKeys caps the number of tracked rate limit keys.
	// If zero or less, the key count is unlimited.
	MaxKeys int
}

// RateLimitMiddleware limits requests by client IP.
func RateLimitMiddleware(limit int, window time.Duration) goapi.HandleFunc {
	return RateLimitMiddlewareWithConfig(RateLimitConfig{
		Limit:  limit,
		Window: window,
	})
}

// RateLimitMiddlewareWithConfig returns a token-bucket rate limit middleware.
func RateLimitMiddlewareWithConfig(config RateLimitConfig) goapi.HandleFunc {
	limiter := newRateLimiter(config)
	keyFunc := config.KeyFunc
	message := config.Message
	if message == "" {
		message = defaultRateLimitMessage
	}

	return func(ctx *goapi.Context) {
		key := ""
		if keyFunc != nil {
			key = keyFunc(ctx)
		}
		if key == "" {
			key = defaultRateLimitKey(ctx)
		}

		allowed, retryAfter := limiter.allow(key, time.Now())
		if allowed {
			ctx.Next()
			return
		}

		if retryAfter > 0 {
			ctx.Writer.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
		}
		http.Error(ctx.Writer, message, http.StatusTooManyRequests)
	}
}

type rateLimiter struct {
	mu              sync.Mutex
	limit           float64
	window          time.Duration
	burst           float64
	buckets         map[string]*rateLimitBucket
	cleanupInterval time.Duration
	nextCleanup     time.Time
	maxKeys         int
}

type rateLimitBucket struct {
	tokens float64
	last   time.Time
	seen   time.Time
}

func newRateLimiter(config RateLimitConfig) *rateLimiter {
	if config.Limit <= 0 {
		panic("goapi: rate limit Limit must be greater than zero")
	}
	if config.Window <= 0 {
		panic("goapi: rate limit Window must be greater than zero")
	}
	burst := config.Burst
	if burst <= 0 {
		burst = config.Limit
	}
	cleanupInterval := config.CleanupInterval
	if cleanupInterval <= 0 {
		cleanupInterval = config.Window
		if cleanupInterval < time.Minute {
			cleanupInterval = time.Minute
		}
	}
	return &rateLimiter{
		limit:           float64(config.Limit),
		window:          config.Window,
		burst:           float64(burst),
		buckets:         make(map[string]*rateLimitBucket),
		cleanupInterval: cleanupInterval,
		maxKeys:         config.MaxKeys,
	}
}

func (l *rateLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	if key == "" {
		key = "global"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.nextCleanup.IsZero() {
		l.nextCleanup = now.Add(l.cleanupInterval)
	} else if !now.Before(l.nextCleanup) {
		l.cleanup(now)
		l.nextCleanup = now.Add(l.cleanupInterval)
	}

	bucket := l.buckets[key]
	if bucket == nil {
		if l.maxKeys > 0 && len(l.buckets) >= l.maxKeys {
			l.cleanup(now)
			if len(l.buckets) >= l.maxKeys {
				return false, l.cleanupInterval
			}
		}
		bucket = &rateLimitBucket{
			tokens: l.burst,
			last:   now,
			seen:   now,
		}
		l.buckets[key] = bucket
	}

	l.refill(bucket, now)
	bucket.seen = now
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true, 0
	}

	return false, l.retryAfter(bucket.tokens)
}

func (l *rateLimiter) refill(bucket *rateLimitBucket, now time.Time) {
	elapsed := now.Sub(bucket.last)
	if elapsed < 0 {
		elapsed = 0
	}
	bucket.last = now
	if elapsed == 0 {
		return
	}

	tokens := bucket.tokens + elapsed.Seconds()*l.limit/l.window.Seconds()
	if tokens > l.burst {
		tokens = l.burst
	}
	bucket.tokens = tokens
}

func (l *rateLimiter) retryAfter(tokens float64) time.Duration {
	seconds := (1 - tokens) * l.window.Seconds() / l.limit
	if seconds <= 0 {
		return 0
	}
	return time.Duration(math.Ceil(seconds * float64(time.Second)))
}

func (l *rateLimiter) cleanup(now time.Time) {
	maxIdle := l.cleanupInterval
	if maxIdle < l.window {
		maxIdle = l.window
	}
	for key, bucket := range l.buckets {
		if now.Sub(bucket.seen) > maxIdle {
			delete(l.buckets, key)
		}
	}
}

func defaultRateLimitKey(ctx *goapi.Context) string {
	if ctx == nil || ctx.Request == nil {
		return "global"
	}
	if ip := ctx.RemoteIP(); ip != "" {
		return ip
	}
	if ctx.Request.RemoteAddr != "" {
		return ctx.Request.RemoteAddr
	}
	return "global"
}
