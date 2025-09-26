package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/josedacruz/architecture-design-system/note-taker-server/internal/note"
	"github.com/josedacruz/architecture-design-system/note-taker-server/internal/note/storage"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteNoteStorage implements the storage.NoteStorage interface using a SQLite database.
type SQLiteNoteStorage struct {
	db *sql.DB // The database connection pool
}

// NewSQLiteNoteStorage creates a new SQLiteNoteStorage instance,
// opens the database connection, and initializes the notes table if it doesn't exist.
func NewSQLiteNoteStorage(dbPath string) (storage.NoteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Ping the database to ensure the connection is established.
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	// Create the notes table if it doesn't already exist.
	// Tags are stored as a JSON string. ReminderTime, CreatedAt, UpdatedAt as TEXT (ISO 8601).
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS notes (
		id TEXT PRIMARY KEY,
		title TEXT,
		content TEXT,
		tags TEXT,          -- Stored as JSON array string
		pinned INTEGER,     -- 0 for false, 1 for true
		archived INTEGER,   -- 0 for false, 1 for true
		reminder_time TEXT, -- Stored as ISO 8601 string (UTC)
		created_at TEXT,    -- Stored as ISO 8601 string (UTC)
		updated_at TEXT     -- Stored as ISO 8601 string (UTC)
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create notes table: %w", err)
	}

	log.Printf("SQLite table 'notes' ensured to exist.")
	return &SQLiteNoteStorage{db: db}, nil
}

// Close closes the underlying SQLite database connection.
func (s *SQLiteNoteStorage) Close() error {
	return s.db.Close()
}

// CreateNote inserts a new note into the SQLite database.
func (s *SQLiteNoteStorage) CreateNote(note note.Note) error {
	tagsJSON, err := json.Marshal(note.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags to JSON: %w", err)
	}

	// Prepare the reminder_time string. If nil, store as NULL.
	var reminderTimeStr sql.NullString
	if note.ReminderTime != nil {
		reminderTimeStr = sql.NullString{String: note.ReminderTime.UTC().Format(time.RFC3339), Valid: true}
	}

	insertSQL := `
	INSERT INTO notes (id, title, content, tags, pinned, archived, reminder_time, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`

	_, err = s.db.Exec(
		insertSQL,
		note.ID,
		note.Title,
		note.Content,
		string(tagsJSON),
		note.Pinned,
		note.Archived,
		reminderTimeStr,
		note.CreatedAt.UTC().Format(time.RFC3339),
		note.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to create note in SQLite: %w", err)
	}
	return nil
}

// GetNoteByID retrieves a note by its ID from the SQLite database.
func (s *SQLiteNoteStorage) GetNoteByID(id string) (note.Note, bool) {
	querySQL := `
	SELECT id, title, content, tags, pinned, archived, reminder_time, created_at, updated_at
	FROM notes WHERE id = ?;`

	row := s.db.QueryRow(querySQL, id)

	var n note.Note
	var tagsJSON string
	var reminderTimeStr sql.NullString
	var createdAtStr, updatedAtStr string

	err := row.Scan(
		&n.ID,
		&n.Title,
		&n.Content,
		&tagsJSON,
		&n.Pinned,
		&n.Archived,
		&reminderTimeStr,
		&createdAtStr,
		&updatedAtStr,
	)

	if err == sql.ErrNoRows {
		return note.Note{}, false // Note not found
	}
	if err != nil {
		log.Printf("Error scanning note from SQLite: %v", err)
		return note.Note{}, false // Other database error
	}

	// Unmarshal tags from JSON string
	if err := json.Unmarshal([]byte(tagsJSON), &n.Tags); err != nil {
		log.Printf("Error unmarshaling tags for note %s: %v", n.ID, err)
		n.Tags = []string{} // Default to empty if unmarshal fails
	}

	// Parse reminder_time if it's not NULL
	if reminderTimeStr.Valid {
		parsedTime, parseErr := time.Parse(time.RFC3339, reminderTimeStr.String)
		if parseErr != nil {
			log.Printf("Error parsing reminder_time for note %s: %v", n.ID, parseErr)
		} else {
			n.ReminderTime = &parsedTime
		}
	}

	// Parse CreatedAt and UpdatedAt
	n.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		log.Printf("Error parsing created_at for note %s: %v", n.ID, err)
	}
	n.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		log.Printf("Error parsing updated_at for note %s: %v", n.ID, err)
	}

	return n, true
}

// GetAllNotes retrieves all notes from the SQLite database.
func (s *SQLiteNoteStorage) GetAllNotes() ([]note.Note, error) {
	querySQL := `
	SELECT id, title, content, tags, pinned, archived, reminder_time, created_at, updated_at
	FROM notes;`

	rows, err := s.db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query all notes from SQLite: %w", err)
	}
	defer rows.Close()

	var notes []note.Note
	for rows.Next() {
		var n note.Note
		var tagsJSON string
		var reminderTimeStr sql.NullString
		var createdAtStr, updatedAtStr string

		if err := rows.Scan(
			&n.ID,
			&n.Title,
			&n.Content,
			&tagsJSON,
			&n.Pinned,
			&n.Archived,
			&reminderTimeStr,
			&createdAtStr,
			&updatedAtStr,
		); err != nil {
			log.Printf("Error scanning row for GetAllNotes: %v", err)
			continue // Skip this row and try next
		}

		// Unmarshal tags
		if err := json.Unmarshal([]byte(tagsJSON), &n.Tags); err != nil {
			log.Printf("Error unmarshaling tags for note %s (GetAllNotes): %v", n.ID, err)
			n.Tags = []string{}
		}

		// Parse reminder_time
		if reminderTimeStr.Valid {
			parsedTime, parseErr := time.Parse(time.RFC3339, reminderTimeStr.String)
			if parseErr != nil {
				log.Printf("Error parsing reminder_time for note %s (GetAllNotes): %v", n.ID, parseErr)
			} else {
				n.ReminderTime = &parsedTime
			}
		}

		// Parse CreatedAt and UpdatedAt
		n.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Printf("Error parsing created_at for note %s (GetAllNotes): %v", n.ID, err)
		}
		n.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			log.Printf("Error parsing updated_at for note %s (GetAllNotes): %v", n.ID, err)
		}

		notes = append(notes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows for GetAllNotes: %w", err)
	}

	return notes, nil
}

// UpdateNote updates an existing note in the SQLite database.
func (s *SQLiteNoteStorage) UpdateNote(note note.Note) error {
	tagsJSON, err := json.Marshal(note.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags to JSON for update: %w", err)
	}

	var reminderTimeStr sql.NullString
	if note.ReminderTime != nil {
		reminderTimeStr = sql.NullString{String: note.ReminderTime.UTC().Format(time.RFC3339), Valid: true}
	}

	updateSQL := `
	UPDATE notes
	SET title = ?, content = ?, tags = ?, pinned = ?, archived = ?, reminder_time = ?, updated_at = ?
	WHERE id = ?;`

	res, err := s.db.Exec(
		updateSQL,
		note.Title,
		note.Content,
		string(tagsJSON),
		note.Pinned,
		note.Archived,
		reminderTimeStr,
		note.UpdatedAt.UTC().Format(time.RFC3339),
		note.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update note in SQLite: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after update: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("note not found for update") // No rows were updated
	}
	return nil
}

// DeleteNote removes a note from the SQLite database by its ID.
func (s *SQLiteNoteStorage) DeleteNote(id string) error {
	deleteSQL := `DELETE FROM notes WHERE id = ?;`

	res, err := s.db.Exec(deleteSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete note from SQLite: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after delete: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("note not found for deletion") // No rows were deleted
	}
	return nil
}
