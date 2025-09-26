package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux" // Popular router for Go HTTP servers
	"github.com/josedacruz/architecture-design-system/note-taker-server/internal/constants"
	"github.com/josedacruz/architecture-design-system/note-taker-server/internal/note"
	"github.com/josedacruz/architecture-design-system/note-taker-server/internal/note/storage/sqlite"
	"github.com/josedacruz/architecture-design-system/note-taker-server/internal/reminder"
)

func main() {

	// 1. Initialize Storage (SQLite for persistence)
	noteStorage, err := sqlite.NewSQLiteNoteStorage(constants.SqliteDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite storage: %v", err)
	}

	// Ensure the database connection is closed when the main function exits.
	defer func() {
		if err := noteStorage.Close(); err != nil {
			log.Printf("Error closing SQLite database: %v", err)
		}
	}()
	log.Printf("SQLite database initialized at %s", constants.SqliteDBPath)

	// 2. Initialize Service Layer
	noteService := note.NewNoteService(noteStorage)

	// 3. Initialize Handler Layer (HTTP API)
	noteHandler := note.NewNoteHandler(noteService)

	// 4. Setup HTTP Router (using gorilla/mux for more advanced routing)
	router := mux.NewRouter()

	// Define API routes for notes (CRUD operations)
	// GET /notes?tags=tag1,tag2 - Get notes by tags. This route must come BEFORE the generic /notes GET.
	// Mux's matching order matters for overlapping routes.
	router.HandleFunc("/notes", noteHandler.GetNotesByTagsHandler).Methods("GET").Queries("tags", "{tags}")
	// GET /notes - Get all notes (if no 'tags' query parameter is present)
	router.HandleFunc("/notes", noteHandler.GetAllNotesHandler).Methods("GET")
	router.HandleFunc("/notes", noteHandler.CreateNoteHandler).Methods("POST")
	router.HandleFunc("/notes/{id}", noteHandler.GetNoteHandler).Methods("GET")
	router.HandleFunc("/notes/{id}", noteHandler.UpdateNoteHandler).Methods("PUT")
	router.HandleFunc("/notes/{id}", noteHandler.DeleteNoteHandler).Methods("DELETE")

	// 5. Initialize and Start Reminder Processor
	reminderProcessor := reminder.NewReminderProcessor(
		noteService, // The noteService instance acts as the noteFetcher
		constants.ReminderCheckInterval,
		constants.ReminderNotificationWindow,
	)

	// Start the reminder processor in a new goroutine so it runs concurrently
	// with the HTTP server without blocking it.
	go reminderProcessor.Start()

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = constants.DefaultPort
	}

	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Starting Note-Taking App API server on %s", serverAddr)
	log.Println("Available Endpoints:")
	log.Println("  POST /notes - Create a new note (with optional reminder_time)")
	log.Println("  GET /notes - Get all notes")
	log.Println("  GET /notes?tags=tag1,tag2 - Get notes filtered by tags")
	log.Println("  GET /notes/{id} - Get a note by ID")
	log.Println("  PUT /notes/{id} - Update a note by ID (with optional reminder_time)")
	log.Println("  DELETE /notes/{id} - Delete a note by ID")

	// Configure HTTP server with timeouts for robustness
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,            // Use the gorilla/mux router to handle requests
		ReadTimeout:  5 * time.Second,   // Max time to read the entire request, including the body.
		WriteTimeout: 10 * time.Second,  // Max time to write the response.
		IdleTimeout:  120 * time.Second, // Max time for a client connection to remain idle.
	}

	// ListenAndServe starts the HTTP server. It blocks until the server stops or an error occurs.
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err) // Log fatal error if server fails
	}
}
