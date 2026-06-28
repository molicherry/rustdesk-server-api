package middleware

import (
	"sync"
	"time"
)

// RateLimiter tracks failed login attempts per key (typically client IP)
// using a sliding window. It supports a hard block threshold (maxAttempts)
// and a softer captcha threshold that signals the client to solve a captcha.
type RateLimiter struct {
	mu               sync.Mutex
	attempts         map[string][]time.Time
	window           time.Duration
	maxAttempts      int
	captchaThreshold int
}

// NewRateLimiter creates a new RateLimiter.
//   - window: sliding window duration for counting attempts
//   - maxAttempts: hard block threshold (return 429 when reached)
//   - captchaThreshold: captcha-required threshold (≤ maxAttempts)
func NewRateLimiter(window time.Duration, maxAttempts int, captchaThreshold int) *RateLimiter {
	return &RateLimiter{
		attempts:         make(map[string][]time.Time),
		window:           window,
		maxAttempts:      maxAttempts,
		captchaThreshold: captchaThreshold,
	}
}

// attemptCount returns the number of attempts still in the window for key.
// Caller must hold rl.mu.
func (rl *RateLimiter) attemptCount(key string) int {
	now := time.Now()
	cutoff := now.Add(-rl.window)
	times := rl.attempts[key]

	// Filter out expired entries in-place.
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.attempts[key] = valid
	return len(valid)
}

// IsBlocked returns true when key has ≥ maxAttempts attempts in the window.
func (rl *RateLimiter) IsBlocked(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.attemptCount(key) >= rl.maxAttempts
}

// CaptchaRequired returns true when key has ≥ captchaThreshold attempts.
func (rl *RateLimiter) CaptchaRequired(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.attemptCount(key) >= rl.captchaThreshold
}

// RecordAttempt records a failed attempt for key.
func (rl *RateLimiter) RecordAttempt(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.attempts[key] = append(rl.attempts[key], time.Now())
}

// Clear removes all attempts for key (called on successful login).
func (rl *RateLimiter) Clear(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

// LoginLimiter is the package-level login rate limiter.
// Initialized during server startup via InitRateLimiter.
var LoginLimiter *RateLimiter

// InitRateLimiter creates the global LoginLimiter.
// captchaThreshold is read from config (app.captcha_threshold, default 3).
func InitRateLimiter(captchaThreshold int) {
	// 10 attempts per 5-minute window before hard block.
	LoginLimiter = NewRateLimiter(5*time.Minute, 10, captchaThreshold)
}
