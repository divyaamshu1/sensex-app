package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"sensex-backend/internal/cache"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	cache *cache.Store
}

// New creates a new Handler
func New(c *cache.Store) *Handler {
	return &Handler{cache: c}
}

// RegisterRoutes wires up all HTTP routes onto the given mux
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/sensex", h.handleSensex)
	mux.HandleFunc("/api/sensex/stream", h.handleSensexStream)
	mux.HandleFunc("/api/health", h.handleHealth)
	mux.HandleFunc("/", h.handleRoot)
}

// handleSensex returns the latest Sensex snapshot as JSON (REST polling)
// GET /api/sensex
func (h *Handler) handleSensex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	snap := h.cache.Get()
	if snap == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Data not yet available, please retry in a moment",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(snap); err != nil {
		log.Printf("[api] JSON encode error: %v", err)
	}
}

// handleSensexStream streams Sensex updates via Server-Sent Events (SSE)
// GET /api/sensex/stream
// The frontend can use EventSource to receive real-time updates
func (h *Handler) handleSensexStream(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	log.Printf("[api/sse] Client connected: %s", r.RemoteAddr)

	// Send a comment to establish the connection immediately
	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Track last update to avoid sending duplicate data
	var lastUpdated string

	for {
		select {
		case <-r.Context().Done():
			log.Printf("[api/sse] Client disconnected: %s", r.RemoteAddr)
			return

		case <-ticker.C:
			snap := h.cache.Get()
			if snap == nil {
				fmt.Fprintf(w, "event: waiting\ndata: {\"message\":\"Fetching data...\"}\n\n")
				flusher.Flush()
				continue
			}

			// Only send if data has changed
			if snap.LastUpdated == lastUpdated {
				// Send a keepalive comment every tick to prevent timeout
				fmt.Fprintf(w, ": keepalive\n\n")
				flusher.Flush()
				continue
			}

			lastUpdated = snap.LastUpdated

			data, err := json.Marshal(snap)
			if err != nil {
				log.Printf("[api/sse] Marshal error: %v", err)
				continue
			}

			fmt.Fprintf(w, "event: sensex\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// handleHealth returns a simple health check response
// GET /api/health
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	snap := h.cache.Get()
	status := "ok"
	lastUpdated := ""

	if snap == nil {
		status = "initializing"
	} else {
		lastUpdated = snap.LastUpdated
		if snap.Error != "" {
			status = "degraded"
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":      status,
		"lastUpdated": lastUpdated,
		"service":     "sensex-backend",
	})
}

// handleRoot redirects / to the API info
func (h *Handler) handleRoot(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "Sensex Backend API",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"GET /api/sensex":        "Latest Sensex snapshot (JSON, poll this)",
			"GET /api/sensex/stream": "Real-time Sensex updates (SSE stream)",
			"GET /api/health":        "Health check",
		},
	})
}

// setCORSHeaders adds CORS headers to allow React frontend to call this API
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
