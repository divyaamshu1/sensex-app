package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sensex-backend/internal/api"
	"sensex-backend/internal/cache"
	"sensex-backend/internal/fetcher"
)

func main() {
	// Command-line flags
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	// Allow PORT env var override (useful for Android)
	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("=== Sensex Backend Starting on port %s ===", *port)

	// Initialize components
	store := cache.NewStore()
	f := fetcher.New()
	poller := fetcher.NewPoller(f, store)
	handler := api.New(store)

	// Start background polling
	poller.Start()

	// Set up HTTP server
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := "0.0.0.0:" + *port
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // 0 = no timeout for SSE streams
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[server] Listening on http://%s", addr)
		log.Printf("[server] Endpoints:")
		log.Printf("[server]   GET http://localhost:%s/api/sensex        — Latest snapshot", *port)
		log.Printf("[server]   GET http://localhost:%s/api/sensex/stream — SSE stream", *port)
		log.Printf("[server]   GET http://localhost:%s/api/health        — Health check", *port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[server] Fatal error: %v", err)
		}
	}()

	<-stop
	log.Println("[server] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[server] Shutdown error: %v", err)
	}

	log.Println("[server] Stopped.")
}
