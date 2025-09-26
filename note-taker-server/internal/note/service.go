package note

import (
	"errors"
	"fmt"
	"strings" // For string manipulation (e.g., checking tags)
	"time"

	"github.com/josedacruz/architecture-design-system/note-taker-server/internal/note/storage"
)

// ReminderNoteFetcher defines an interface for services that can fetch notes
// for reminder processing. This promotes loose coupling, allowing the ReminderProcessor
// to only depend on what it needs from the NoteService.
type ReminderNoteFetcher interface {
	GetAllNotes() ([]Note, error) // Returns internal Note struct, not NoteResponse DTO
}

// NoteService defines the business logic operations for notes.
// This interface exposes the functionalities that the handlers will use.
type NoteService interface {
	CreateNote(req CreateNoteRequest) (NoteResponse, error)
	GetNoteByID(id string) (NoteResponse, error)
	GetAllNotes() ([]NoteResponse, error)
	GetNotesByTags(tags []string) ([]NoteResponse, error) // Method for filtering by tags
	UpdateNote(id string, req UpdateNoteRequest) (NoteResponse, error)
	DeleteNote(id string) error
	// GetAllNotesForReminderProcessor is a specific method for the reminder processor
	// to fetch all notes. It returns the internal Note struct, not the DTO.
	GetAllNotesForReminderProcessor() ([]Note, error)
}

// noteService implements the NoteService interface.
// It orchestrates operations by interacting with the NoteStorage.
type noteService struct {
	storage storage.NoteStorage // Dependency on the storage layer (now uses the interface)
}

// NewNoteService creates and returns a new noteService instance.
// It takes a NoteStorage implementation as a dependency.
func NewNoteService(s storage.NoteStorage) NoteService {
	return &noteService{
		storage: s,
	}
}

// CreateNote handles the creation of a new note.
// It validates the input, generates a unique ID, sets timestamps,
// and then delegates to the storage layer to save the note.
func (s *noteService) CreateNote(req CreateNoteRequest) (NoteResponse, error) {
	// Basic validation: A note should have at least a title or content.
	if req.Title == "" && req.Content == "" {
		return NoteResponse{}, errors.New("note must have a title or content")
	}

	now := time.Now() // Get the current time for creation and update timestamps
	newNote := Note{
		ID:           util.GenerateUUID(), // Generate a globally unique ID for the new note
		Title:        req.Title,
		Content:      req.Content,
		Tags:         req.Tags,
		Pinned:       false,            // Default to not pinned
		Archived:     false,            // Default to not archived
		ReminderTime: req.ReminderTime, // Set reminder time if provided by the client
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Attempt to save the new note to storage.
	err := s.storage.CreateNote(newNote)
	if err != nil {
		// Wrap the error from storage for better context.
		return NoteResponse{}, fmt.Errorf("failed to create note in storage: %w", err)
	}

	// Convert the internal Note model to a NoteResponse DTO before returning.
	return toNoteResponse(newNote), nil
}

// GetNoteByID retrieves a note by its ID.
// It delegates to the storage layer and then converts the result to a DTO.
func (s *noteService) GetNoteByID(id string) (NoteResponse, error) {
	note, found := s.storage.GetNoteByID(id)
	if !found {
		return NoteResponse{}, errors.New("note not found")
	}
	return toNoteResponse(note), nil
}

// GetAllNotes retrieves all notes from storage.
// It delegates to the storage layer and converts all retrieved notes to DTOs.
func (s *noteService) GetAllNotes() ([]NoteResponse, error) {
	notes, err := s.storage.GetAllNotes()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all notes from storage: %w", err)
	}

	// Convert each internal Note model to a NoteResponse DTO.
	responses := make([]NoteResponse, len(notes))
	for i, note := range notes {
		responses[i] = toNoteResponse(note)
	}
	return responses, nil
}

// GetNotesByTags retrieves notes that contain at least one of the specified tags.
// This method filters notes after retrieving all of them from storage.
func (s *noteService) GetNotesByTags(tags []string) ([]NoteResponse, error) {
	allNotes, err := s.storage.GetAllNotes()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all notes for tag filtering: %w", err)
	}

	filteredNotes := make([]NoteResponse, 0)
	tagSet := make(map[string]struct{}) // Use a set for efficient tag lookup
	for _, tag := range tags {
		tagSet[strings.ToLower(tag)] = struct{}{} // Store tags in lowercase for case-insensitive matching
	}

	for _, note := range allNotes {
		for _, noteTag := range note.Tags {
			if _, found := tagSet[strings.ToLower(noteTag)]; found {
				filteredNotes = append(filteredNotes, toNoteResponse(note))
				break // Found a matching tag, move to the next note
			}
		}
	}
	return filteredNotes, nil
}

// UpdateNote updates an existing note.
// It fetches the existing note, applies updates from the request,
// updates the timestamp, and then delegates to the storage layer.
func (s *noteService) UpdateNote(id string, req UpdateNoteRequest) (NoteResponse, error) {
	existingNote, found := s.storage.GetNoteByID(id)
	if !found {
		return NoteResponse{}, errors.New("note not found")
	}

	// Apply updates from the request. Pointers in UpdateNoteRequest allow
	// clients to send only the fields they want to change (omitempty).
	if req.Title != nil {
		existingNote.Title = *req.Title
	}
	if req.Content != nil {
		existingNote.Content = *req.Content
	}
	if req.Tags != nil {
		existingNote.Tags = req.Tags
	}
	if req.Pinned != nil {
		existingNote.Pinned = *req.Pinned
	}
	if req.Archived != nil {
		existingNote.Archived = *req.Archived
	}
	// Handle ReminderTime update:
	// If req.ReminderTime is not nil, it means the client explicitly sent a value for reminder_time.
	// This value itself could be a pointer to a time.Time or a nil pointer (to clear the reminder).
	if req.ReminderTime != nil {
		existingNote.ReminderTime = *req.ReminderTime // Dereference the outer pointer
	}

	existingNote.UpdatedAt = time.Now() // Update the timestamp of the last modification

	// Attempt to update the note in storage.
	err := s.storage.UpdateNote(existingNote)
	if err != nil {
		return NoteResponse{}, fmt.Errorf("failed to update note in storage: %w", err)
	}

	return toNoteResponse(existingNote), nil
}

// DeleteNote deletes a note by its ID.
// It simply delegates the deletion to the storage layer.
func (s *noteService) DeleteNote(id string) error {
	err := s.storage.DeleteNote(id)
	if err != nil {
		return fmt.Errorf("failed to delete note from storage: %w", err)
	}
	return nil
}

// GetAllNotesForReminderProcessor is a specific method implemented for the
// ReminderNoteFetcher interface. It allows the ReminderProcessor to get
// all notes in their internal `Note` struct format, which is suitable
// for internal processing without DTO conversion overhead.
func (s *noteService) GetAllNotesForReminderProcessor() ([]Note, error) {
	return s.storage.GetAllNotes()
}

// toNoteResponse converts an internal `Note` model struct to a `NoteResponse` DTO.
// This function helps in separating internal data representation from external API responses.
func toNoteResponse(note Note) NoteResponse {
	return NoteResponse{
		ID:           note.ID,
		Title:        note.Title,
		Content:      note.Content,
		Tags:         note.Tags,
		Pinned:       note.Pinned,
		Archived:     note.Archived,
		ReminderTime: note.ReminderTime,
		CreatedAt:    note.CreatedAt,
		UpdatedAt:    note.UpdatedAt,
	}
}
