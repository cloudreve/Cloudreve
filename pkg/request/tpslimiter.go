package request

import (
	"context"
	"golang.org/x/time/rate"
	"sync"
)

var globalTPSLimiter = NewTPSLimiter()

type TPSLimiter interface {
	Limit(ctx context.Context, token string, tps float64, burst int)
}

func NewTPSLimiter() TPSLimiter {
	return &multipleBucketLimiter{
		buckets: make(map[string]*rate.Limiter),
	}
}

// multipleBucketLimiter implements TPSLimiter with multiple bucket support.
type multipleBucketLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rate.Limiter
}

// Limit finds the given bucket, if bucket not exist or limit is changed,
// a new bucket will be generated.
func (m *multipleBucketLimiter) Limit(ctx context.Context, token string, tps float64, burst int) {
	m.mu.Lock()
	bucket, ok := m.buckets[token]
	if !ok || float64(bucket.Limit()) != tps || bucket.Burst() != burst {
		bucket = rate.NewLimiter(rate.Limit(tps), burst)
		m.buckets[token] = bucket
	}
	m.mu.Unlock()

	bucket.Wait(ctx)
}
