package note

import (
	"encoding/json" // For marshaling/unmarshaling JSON
	"net/http"      // For HTTP server functionalities
	"strings"       // For string manipulation (e.g., splitting tags)

	"github.com/gorilla/mux" // Popular router for Go HTTP servers, used for path variables
)

// NoteHandler provides HTTP handlers for the Note API.
// It holds a reference to the NoteService to interact with the business logic.
type NoteHandler struct {
	service NoteService // Dependency on the NoteService interface
}

// NewNoteHandler creates and returns a new NoteHandler instance.
func NewNoteHandler(s NoteService) *NoteHandler {
	return &NoteHandler{
		service: s,
	}
}

// respondWithJSON is a helper function to send JSON responses with a specified HTTP status code.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Marshal the payload (any Go data structure) into a JSON byte slice.
	response, err := json.Marshal(payload)
	if err != nil {
		// If marshaling fails, send a generic internal server error.
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json") // Set Content-Type header
	w.WriteHeader(code)                                // Write the HTTP status code
	w.Write(response)                                  // Write the JSON response body
}

// respondWithError is a helper function to send JSON error responses.
func respondWithError(w http.ResponseWriter, code int, message string) {
	// Use respondWithJSON to send an ErrorResponse struct.
	respondWithJSON(w, code, ErrorResponse{Message: message})
}

// CreateNoteHandler handles POST requests to create a new note.
// Route: POST /notes
func (h *NoteHandler) CreateNoteHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateNoteRequest
	// Decode the JSON request body into the CreateNoteRequest struct.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Call the service layer to create the note.
	note, err := h.service.CreateNote(req)
	if err != nil {
		// Handle service-level errors and return appropriate HTTP status codes.
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// On success, respond with 201 Created and the created note's details.
	respondWithJSON(w, http.StatusCreated, note)
}

// GetNoteHandler handles GET requests to retrieve a single note by ID.
// Route: GET /notes/{id}
func (h *NoteHandler) GetNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // Get path variables from the request (e.g., "id")
	id := vars["id"]    // Extract the note ID from the path

	// Call the service layer to get the note by ID.
	note, err := h.service.GetNoteByID(id)
	if err != nil {
		// Handle specific "not found" error.
		if err.Error() == "note not found" {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		// For other errors, return internal server error.
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// On success, respond with 200 OK and the note's details.
	respondWithJSON(w, http.StatusOK, note)
}

// GetAllNotesHandler handles GET requests to retrieve all notes.
// Route: GET /notes (when no 'tags' query parameter is present)
func (h *NoteHandler) GetAllNotesHandler(w http.ResponseWriter, r *http.Request) {
	// Call the service layer to get all notes.
	notes, err := h.service.GetAllNotes()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// On success, respond with 200 OK and the list of notes.
	respondWithJSON(w, http.StatusOK, notes)
}

// GetNotesByTagsHandler handles GET requests to retrieve notes filtered by tags.
// Route: GET /notes?tags=tag1,tag2
func (h *NoteHandler) GetNotesByTagsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the 'tags' query parameter from the URL.
	// The Mux router ensures this handler is only called if 'tags' is present.
	tagsParam := r.URL.Query().Get("tags")
	if tagsParam == "" {
		// This case should ideally not be hit if router is configured correctly
		// with .Queries("tags", "{tags}"). But as a fallback/safety, handle it.
		respondWithError(w, http.StatusBadRequest, "Tags query parameter is required")
		return
	}

	// Split the comma-separated tags string into a slice of strings.
	tags := strings.Split(tagsParam, ",")

	// Call the service layer to get notes filtered by these tags.
	notes, err := h.service.GetNotesByTags(tags)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// On success, respond with 200 OK and the filtered list of notes.
	respondWithJSON(w, http.StatusOK, notes)
}

// UpdateNoteHandler handles PUT requests to update an existing note.
// Route: PUT /notes/{id}
func (h *NoteHandler) UpdateNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"] // Extract the note ID from the path

	var req UpdateNoteRequest
	// Decode the JSON request body into the UpdateNoteRequest struct.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// Call the service layer to update the note.
	note, err := h.service.UpdateNote(id, req)
	if err != nil {
		// Handle specific "not found" error.
		if err.Error() == "note not found" {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		// For other errors, return internal server error.
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// On success, respond with 200 OK and the updated note's details.
	respondWithJSON(w, http.StatusOK, note)
}

// DeleteNoteHandler handles DELETE requests to delete a note by ID.
// Route: DELETE /notes/{id}
func (h *NoteHandler) DeleteNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"] // Extract the note ID from the path

	// Call the service layer to delete the note.
	err := h.service.DeleteNote(id)
	if err != nil {
		// Handle specific "not found" error from the storage layer.
		if err.Error() == "note not found for deletion" {
			respondWithError(w, http.StatusNotFound, "Note not found")
			return
		}
		// For other errors, return internal server error.
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// On successful deletion, respond with 204 No Content.
	w.WriteHeader(http.StatusNoContent)
}
