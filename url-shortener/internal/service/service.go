package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/josedacruz/architecture-design-system/url-shortener/internal/storage"
	"github.com/josedacruz/architecture-design-system/url-shortener/pkg/base62"
)

// ServiceInterface defines the contract for the URL shortener business logic.
// This abstraction allows for different implementations of the core logic.
type ServiceInterface interface {
	ShortenURL(longURL string) (string, error)   // Shortens a given long URL
	GetLongURL(shortCode string) (string, error) // Retrieves the original long URL for a short code
}

// Service implements the ServiceInterface, providing the business logic for URL shortening.
type Service struct {
	storage   storage.Storage // The underlying storage mechanism (e.g., InMemoryStorage, or a database client)
	counterMu sync.Mutex
	counter   int64
}

// NewService creates and returns a new Service instance.
func NewService(s storage.Storage) *Service {
	return &Service{
		storage: s,
	}
}

// ShortenURL generates a short code for a given long URL and stores the mapping.
// It first checks if the long URL has already been shortened to avoid duplicates.
func (s *Service) ShortenURL(longURL string) (string, error) {
	// Check if the long URL has already been shortened. If so, return the existing short code.
	if existingShortCode, ok := s.storage.GetShortCode(longURL); ok {
		return existingShortCode, nil
	}

	// Generate a new unique short code.
	shortCode := s.GenerateShortCode()
	// Save the mapping in the storage.
	err := s.storage.Save(shortCode, longURL)
	if err != nil {
		// If saving fails (e.g., due to a rare collision in a real DB, or other storage errors),
		// return an error.
		return "", fmt.Errorf("failed to save URL mapping: %w", err)
	}
	return shortCode, nil // Return the newly generated short code
}

// GetLongURL retrieves the original long URL for a given short code from storage.
func (s *Service) GetLongURL(shortCode string) (string, error) {
	longURL, ok := s.storage.Get(shortCode) // Attempt to retrieve the long URL
	if !ok {
		// If the short code is not found in storage, return an error.
		return "", errors.New("short code not found")
	}
	return longURL, nil // Return the found long URL
}

// GenerateShortCode increments the internal counter and converts it to a Base62 string.
// This simple counter-based approach guarantees uniqueness for a single instance.
// In a distributed system, a more sophisticated ID generation strategy would be required.
func (s *Service) GenerateShortCode() string {
	s.counterMu.Lock() // Acquire a write lock to safely increment the counter
	defer s.counterMu.Unlock()

	s.counter++                       // Increment the counter
	return base62.ToBase62(s.counter) // Convert the new counter value to Base62
}
