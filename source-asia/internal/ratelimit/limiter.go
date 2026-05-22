package ratelimit

import (
	"sync"
	"time"
)

const (
	WindowDuration = 60 * time.Second
	MaxRequests    = 5
)

type Bucket struct {
	mu            sync.Mutex
	timestamps    []time.Time
	rejectedTotal int
}

type Limiter struct {
	mu      sync.RWMutex
	buckets map[string]*Bucket
}

func NewLimiter() *Limiter {
	return &Limiter{
		buckets: make(map[string]*Bucket),
	}
}

func (l *Limiter) Allow(userID string) bool {
	bucket := l.getOrCreate(userID)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	bucket.timestamps = evictOld(bucket.timestamps, now)

	if len(bucket.timestamps) >= MaxRequests {
		bucket.rejectedTotal++
		return false
	}

	bucket.timestamps = append(bucket.timestamps, now)
	return true
}

func (l *Limiter) Stats(userID string) (int, int) {
	bucket := l.getOrCreate(userID)

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	bucket.timestamps = evictOld(bucket.timestamps, now)

	return len(bucket.timestamps), bucket.rejectedTotal
}

func (l *Limiter) AllUserIDs() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	ids := make([]string, 0, len(l.buckets))
	for id := range l.buckets {
		ids = append(ids, id)
	}
	return ids
}

func (l *Limiter) getOrCreate(userID string) *Bucket {
	l.mu.RLock()
	b, ok := l.buckets[userID]
	l.mu.RUnlock()
	if ok {
		return b
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if b, ok = l.buckets[userID]; ok {
		return b
	}

	b = &Bucket{}
	l.buckets[userID] = b
	return b
}

func evictOld(ts []time.Time, now time.Time) []time.Time {
	cutoff := now.Add(-WindowDuration)
	i := 0
	for i < len(ts) && ts[i].Before(cutoff) {
		i++
	}
	return ts[i:]
}
