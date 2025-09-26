package note

import (
	"time"
)

// Note represents the core structure of a note stored in the system.
// It includes fields for content, metadata, and an optional reminder time.
type Note struct {
	ID           string     `json:"id"`                      // Unique identifier for the note (e.g., UUID)
	Title        string     `json:"title"`                   // Title of the note
	Content      string     `json:"content"`                 // Main content of the note
	Tags         []string   `json:"tags"`                    // List of tags associated with the note
	Pinned       bool       `json:"pinned"`                  // True if the note is pinned to the top
	Archived     bool       `json:"archived"`                // True if the note is archived
	ReminderTime *time.Time `json:"reminder_time,omitempty"` // Optional reminder timestamp. Pointer allows nil.
	CreatedAt    time.Time  `json:"created_at"`              // Timestamp when the note was created
	UpdatedAt    time.Time  `json:"updated_at"`              // Timestamp when the note was last updated
}

// CreateNoteRequest defines the structure for the request body when creating a new note.
// Fields are directly mapped from JSON input.
type CreateNoteRequest struct {
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Tags         []string   `json:"tags"`
	ReminderTime *time.Time `json:"reminder_time,omitempty"` // Optional reminder timestamp from client
}

// UpdateNoteRequest defines the structure for the request body when updating an existing note.
// Pointers are used for fields that can be partially updated (omitted if not provided).
// For ReminderTime, a pointer to a pointer allows distinguishing between:
// - field not sent (outer pointer nil) -> no change to reminder
// - field sent as null (outer pointer non-nil, inner pointer nil) -> clear reminder
// - field sent with a time (outer pointer non-nil, inner pointer non-nil) -> set/update reminder
type UpdateNoteRequest struct {
	Title        *string     `json:"title,omitempty"` // Pointer to allow partial updates
	Content      *string     `json:"content,omitempty"`
	Tags         []string    `json:"tags,omitempty"`
	Pinned       *bool       `json:"pinned,omitempty"`
	Archived     *bool       `json:"archived,omitempty"`
	ReminderTime **time.Time `json:"reminder_time,omitempty"` // Pointer to pointer for flexible update/clear
}

// NoteResponse defines the structure for the response body when returning note details to clients.
// It mirrors the Note struct but serves as a Data Transfer Object (DTO).
type NoteResponse struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Tags         []string   `json:"tags"`
	Pinned       bool       `json:"pinned"`
	Archived     bool       `json:"archived"`
	ReminderTime *time.Time `json:"reminder_time,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ErrorResponse defines a standard format for API error responses.
type ErrorResponse struct {
	Message string `json:"message"` // A descriptive error message
}
