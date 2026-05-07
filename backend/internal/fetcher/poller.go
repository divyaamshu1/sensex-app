package fetcher

import (
	"log"
	"time"

	"sensex-backend/internal/cache"
	"sensex-backend/internal/models"
)

const (
	// How often to refresh Sensex index price (every 1 second)
	indexPollInterval = 1 * time.Second

	// How often to refresh constituent weightages (every 5 minutes - weightages change slowly)
	weightagePollInterval = 5 * time.Minute
)

// Poller continuously refreshes Sensex data in the background
type Poller struct {
	fetcher      *Fetcher
	cache        *cache.Store
	constituents []models.Constituent // cached constituents (refreshed every 5 min)
}

// NewPoller creates a new background poller
func NewPoller(f *Fetcher, c *cache.Store) *Poller {
	return &Poller{
		fetcher: f,
		cache:   c,
	}
}

// Start begins polling in the background
// It returns immediately and runs goroutines in the background
func (p *Poller) Start() {
	log.Println("[poller] Starting background polling...")

	// Initial fetch of constituents (blocking, so first response is ready)
	if err := p.refreshConstituents(); err != nil {
		log.Printf("[poller] Initial constituent fetch failed: %v", err)
	}

	// Initial index fetch
	if err := p.refreshIndex(); err != nil {
		log.Printf("[poller] Initial index fetch failed: %v", err)
	}

	// Background goroutine: refresh index every second
	go p.pollIndex()

	// Background goroutine: refresh constituents every 5 minutes
	go p.pollWeightages()

	log.Println("[poller] Background polling started")
}

// pollIndex refreshes the Sensex index value every second
func (p *Poller) pollIndex() {
	ticker := time.NewTicker(indexPollInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := p.refreshIndex(); err != nil {
			log.Printf("[poller] Index refresh error: %v", err)
		}
	}
}

// pollWeightages refreshes constituent weightages every 5 minutes
func (p *Poller) pollWeightages() {
	ticker := time.NewTicker(weightagePollInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := p.refreshConstituents(); err != nil {
			log.Printf("[poller] Constituent refresh error: %v", err)
		}
	}
}

// refreshIndex fetches the latest Sensex index and builds a snapshot
func (p *Poller) refreshIndex() error {
	idx, err := p.fetcher.FetchSensexIndex()
	if err != nil {
		// Store error in cache so frontend can show it
		existing := p.cache.Get()
		if existing != nil {
			existing.Error = err.Error()
			existing.LastUpdated = time.Now().Format(time.RFC3339)
			p.cache.Set(existing)
		} else {
			p.cache.Set(&models.SensexSnapshot{
				Error:       err.Error(),
				LastUpdated: time.Now().Format(time.RFC3339),
			})
		}
		return err
	}

	// Clone constituents and calculate their point contributions
	constituents := make([]models.Constituent, len(p.constituents))
	copy(constituents, p.constituents)
	constituents = CalculatePointsImpact(constituents, idx.Change)

	snap := &models.SensexSnapshot{
		Index:        *idx,
		Constituents: constituents,
		LastUpdated:  time.Now().Format(time.RFC3339),
	}

	p.cache.Set(snap)
	return nil
}

// refreshConstituents fetches fresh constituent weightages from BSE
func (p *Poller) refreshConstituents() error {
	c, err := p.fetcher.FetchConstituents()
	if err != nil {
		return err
	}
	p.constituents = c
	log.Printf("[poller] Refreshed %d constituents", len(c))
	return nil
}
