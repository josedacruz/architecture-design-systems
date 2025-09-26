package storage

import "github.com/josedacruz/architecture-design-system/note-taker-server/internal/note"

// NoteStorage defines the interface for note data persistence operations.
// Any type that implements these methods can serve as a storage backend for notes.
type NoteStorage interface {
	CreateNote(note note.Note) error         // Adds a new note to storage
	GetNoteByID(id string) (note.Note, bool) // Retrieves a note by its unique ID
	GetAllNotes() ([]note.Note, error)       // Retrieves all notes currently in storage
	UpdateNote(note note.Note) error         // Updates an existing note in storage
	DeleteNote(id string) error              // Deletes a note from storage by its ID
	Close() error                            // Closes the storage connection (e.g., database connection)
}
