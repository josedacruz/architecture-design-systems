package storage

import (
	"errors"
	"sync"
)

// Storage defines the interface for URL mapping storage.
// Any type that implements these methods can be used as storage.
type Storage interface {
	Save(shortCode, longURL string) error       // Stores a shortCode to longURL mapping
	Get(shortCode string) (string, bool)        // Retrieves the long URL for a given shortCode
	GetShortCode(longURL string) (string, bool) // Retrieves the short code for a given long URL (optional, for uniqueness checks)
}

// InMemoryStorage implements the Storage interface using in-memory maps.
// This is suitable for educational purposes but data will be lost on server restart.
type InMemoryStorage struct {
	mu          sync.RWMutex      // A Read-Write Mutex to protect concurrent access to maps
	shortToLong map[string]string // Maps short codes to long URLs
	longToShort map[string]string // Maps long URLs to short codes (for checking if a URL is already shortened)
}

// NewInMemoryStorage creates and returns a new InMemoryStorage instance.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		shortToLong: make(map[string]string), // Initialize the shortCode -> longURL map
		longToShort: make(map[string]string), // Initialize the longURL -> shortCode map
	}
}

// Save stores a shortCode to longURL mapping in the in-memory maps.
// It acquires a write lock to prevent race conditions during write operations.
func (s *InMemoryStorage) Save(shortCode, longURL string) error {
	s.mu.Lock()         // Acquire an exclusive write lock
	defer s.mu.Unlock() // Ensure the lock is released when the function exits

	// Check if shortCode already exists. With the current generation logic, this
	// should ideally not happen unless there's an issue with counter uniqueness
	// or if custom short codes were allowed and conflicted.
	if _, exists := s.shortToLong[shortCode]; exists {
		return errors.New("short code already exists")
	}
	// Check if the longURL has already been shortened to prevent duplicate entries
	// for the same long URL.
	if _, exists := s.longToShort[longURL]; exists {
		return errors.New("long URL already shortened")
	}

	s.shortToLong[shortCode] = longURL // Store the forward mapping
	s.longToShort[longURL] = shortCode // Store the reverse mapping
	return nil
}

// Get retrieves the long URL for a given shortCode from the in-memory map.
// It acquires a read lock, allowing multiple concurrent reads.
func (s *InMemoryStorage) Get(shortCode string) (string, bool) {
	s.mu.RLock()         // Acquire a shared read lock
	defer s.mu.RUnlock() // Ensure the lock is released

	longURL, ok := s.shortToLong[shortCode] // Look up the long URL
	return longURL, ok                      // Return the URL and a boolean indicating if found
}

// GetShortCode retrieves the short code for a given long URL from the in-memory map.
// This is used to check if a long URL has already been shortened.
func (s *InMemoryStorage) GetShortCode(longURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	shortCode, ok := s.longToShort[longURL]
	return shortCode, ok
}
