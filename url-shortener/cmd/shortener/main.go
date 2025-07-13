package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/josedacruz/architecture-design-system/url-shortener/internal/config"
	"github.com/josedacruz/architecture-design-system/url-shortener/internal/handler"
	"github.com/josedacruz/architecture-design-system/url-shortener/internal/service"
	"github.com/josedacruz/architecture-design-system/url-shortener/internal/storage"
)

func main() {
	// 1. Initialize In-Memory Storage
	// This will hold our URL mappings. Data is volatile and will be lost on restart.
	storage := storage.NewInMemoryStorage()

	// 2. Initialize the Core Shortener Service
	// This layer contains the business logic, abstracting away storage details.
	service := service.NewService(storage)

	// 3. Determine the Base URL for Short Links
	// Get from environment variable BASE_URL, or use default.
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = config.DefaultBaseURL
	}
	// Ensure the base URL ends with a slash for correct short URL construction.
	if baseURL[len(baseURL)-1] != '/' {
		baseURL += "/"
	}

	// 4. Initialize HTTP Handlers
	// These handlers connect HTTP requests to our service logic.
	handler := handler.NewHandler(service, baseURL)

	// 5. Register HTTP Routes
	// Map URL paths to their respective handler functions.
	http.HandleFunc("/shorten", handler.Shorten) // Route for shortening URLs (POST)
	http.HandleFunc("/", handler.Redirect)       // Catch-all route for redirecting short codes (GET)

	// 6. Determine the Server Port
	// Get from environment variable PORT, or use default.
	port := os.Getenv("PORT")
	if port == "" {
		port = config.DefaultPort
	}

	// Construct the server address string.
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Starting URL Shortener server on %s", serverAddr)
	log.Printf("Base URL for short links: %s", baseURL)
	log.Printf("Routes:")
	log.Printf("  POST /shorten (to create a short URL)")
	log.Printf("  GET /{shortCode} (to redirect to the original URL)")

	// 7. Configure and Start the HTTP Server
	// Set up server timeouts for better robustness in production.
	srv := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  5 * time.Second,   // Max time to read the entire request, including the body.
		WriteTimeout: 10 * time.Second,  // Max time to write the response.
		IdleTimeout:  120 * time.Second, // Max time for a client connection to remain idle.
	}

	// ListenAndServe starts the HTTP server. It blocks until the server stops or an error occurs.
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err) // Log fatal error if server fails
	}
}
