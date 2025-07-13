package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/josedacruz/architecture-design-system/url-shortener/internal/model"
	"github.com/josedacruz/architecture-design-system/url-shortener/internal/service"
	"github.com/josedacruz/architecture-design-system/url-shortener/pkg/urls"
)

// Handler provides HTTP handlers for the shortener service's API endpoints.
type Handler struct {
	service service.ServiceInterface // The service containing the business logic
	baseURL string                   // The base URL of the shortener service (e.g., "http://localhost:8080/")
}

// NewHandler creates and returns a new Handler instance.
func NewHandler(s service.ServiceInterface, baseURL string) *Handler {
	return &Handler{
		service: s,
		baseURL: baseURL,
	}
}

// Shorten handles POST requests to create a short URL.
// It expects a JSON request body like: `{"long_url": "https://example.com/very/long/url"}`
// It responds with JSON like: `{"short_url": "http://localhost:8080/abcde"}`
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests for this endpoint.
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.ShortenRequest
	// Decode the JSON request body into the ShortenRequest struct.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the provided long URL.
	if req.LongURL == "" || !urls.IsValidURL(req.LongURL) {
		http.Error(w, "Invalid or empty 'long_url' provided. Must be a valid http(s) URL.", http.StatusBadRequest)
		return
	}

	// Call the core service to shorten the URL.
	shortCode, err := h.service.ShortenURL(req.LongURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to shorten URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Construct the full short URL using the base URL and the generated short code.
	resp := model.ShortenResponse{
		ShortURL: h.baseURL + shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Redirect handles GET requests to redirect to the original long URL.
// It expects the short code in the URL path, e.g., `/abcde`.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests for this endpoint.
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the short code from the URL path.
	// r.URL.Path will be like "/abcde", so we slice it to remove the leading "/".
	shortCode := r.URL.Path[1:]

	// If no short code is provided (e.g., request to the root "/"), return a bad request error.
	// In a real application, the root path might serve a landing page.
	if shortCode == "" {
		http.Error(w, "Short code not provided in URL path", http.StatusBadRequest)
		return
	}

	// Call the core service to retrieve the original long URL.
	longURL, err := h.service.GetLongURL(shortCode)
	if err != nil {
		// If the short code is not found, return a 404 Not Found error.
		if err.Error() == "short code not found" {
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}
		// For other service errors, return a 500 Internal Server Error.
		http.Error(w, fmt.Sprintf("Failed to retrieve long URL: %v", err), http.StatusInternalServerError)
		return
	}

	log.Println(shortCode, longURL)

	// Perform a permanent redirect (HTTP 301 Moved Permanently) to the original long URL.
	// This tells the browser to update its bookmarks and cache the new location.
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}
