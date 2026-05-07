package cache

import (
	"sync"
	"time"

	"sensex-backend/internal/models"
)

// Store is a thread-safe cache for the latest SensexSnapshot
type Store struct {
	mu          sync.RWMutex
	snapshot    *models.SensexSnapshot
	lastUpdated time.Time
}

// NewStore creates a new cache store
func NewStore() *Store {
	return &Store{}
}

// Set updates the cached snapshot
func (s *Store) Set(snap *models.SensexSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshot = snap
	s.lastUpdated = time.Now()
}

// Get retrieves the latest cached snapshot (nil if not set)
func (s *Store) Get() *models.SensexSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

// LastUpdated returns when the cache was last refreshed
func (s *Store) LastUpdated() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastUpdated
}
