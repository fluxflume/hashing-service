package http

import (
	"context"
	"fmt"
	"hashing/src/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var createPath = "/hash"
var getPath = fmt.Sprintf("%s/", createPath)
var statsPath = "/stats"
var shutdownPath = "/shutdown"

func StartAndWaitUntilShutdown(port int, hashingService services.HashingService, statsService services.StatsService) {

	// Prepare a channel for handling shutdown/termination requests
	shutdownRequest := make(chan os.Signal, 1)
	signal.Notify(shutdownRequest, syscall.SIGTERM, syscall.SIGINT)

	mux := http.NewServeMux()
	mux.HandleFunc(createPath, recordStats(handleCreate(hashingService), statsService))
	mux.HandleFunc(getPath, handleGet(hashingService))
	mux.HandleFunc(statsPath, handleStats(statsService))
	mux.HandleFunc(shutdownPath, handleShutdown(shutdownRequest))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting server and binding to address %s", addr)
	server := &http.Server{Addr: addr, Handler: mux}

	// In the background, shutdown the server if a shutdown request was published
	go func() {
		<-shutdownRequest
		log.Printf("Shutting down server")
		// Note: Cancel function is not used and probably would never be called in this scenario?
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		if err := server.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server shutdown error: %v", err)
		}
		log.Printf("Server is shutdown")
	}()

	// Start the server and wait until it's shutdown
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}
