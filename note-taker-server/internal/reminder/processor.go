package reminder

import (
	"log"  // For logging notifications
	"sync" // For mutex to protect shared state
	"time" // For time-related operations (tickers, durations)

	"github.com/josedacruz/architecture-design-system/note-taker-server/internal/note"
)

// ReminderProcessor is responsible for periodically checking notes
// and logging notifications for upcoming reminders.
type ReminderProcessor struct {
	noteFetcher        note.ReminderNoteFetcher // Interface to fetch notes (e.g., from NoteService)
	interval           time.Duration            // How often the processor checks for reminders (e.g., 10s)
	notificationWindow time.Duration            // How far in advance to notify for a reminder (e.g., 5m)
	// notifiedReminders tracks notes that have already triggered a notification in the current session.
	// This prevents repeated logging for the same reminder in an in-memory setup.
	// In a persistent system, this state would be stored in a database.
	notifiedReminders map[string]bool
	mu                sync.Mutex // A mutex to protect concurrent access to the `notifiedReminders` map.
}

// NewReminderProcessor creates and returns a new ReminderProcessor instance.
// It takes a `note.ReminderNoteFetcher` (typically the NoteService instance),
// a check interval, and a notification window as parameters.
func NewReminderProcessor(fetcher note.ReminderNoteFetcher, interval time.Duration, notificationWindow time.Duration) *ReminderProcessor {
	return &ReminderProcessor{
		noteFetcher:        fetcher,
		interval:           interval,
		notificationWindow: notificationWindow,
		notifiedReminders:  make(map[string]bool), // Initialize the map
	}
}

// Start begins the reminder checking loop in a new goroutine.
// It uses a `time.Ticker` to trigger the `checkReminders` method at regular intervals.
func (rp *ReminderProcessor) Start() {
	log.Printf("Reminder processor started, checking every %s for reminders due in next %s", rp.interval, rp.notificationWindow)
	ticker := time.NewTicker(rp.interval)
	defer ticker.Stop() // Ensure the ticker is stopped when Start exits, preventing resource leaks

	// Run the first check immediately when the processor starts.
	rp.checkReminders()

	// Loop indefinitely, checking for reminders on each tick of the ticker.
	for range ticker.C {
		rp.checkReminders()
	}
}

// checkReminders fetches all notes from the `noteFetcher` and identifies
// and logs notifications for notes with upcoming reminder times.
func (rp *ReminderProcessor) checkReminders() {
	// Fetch all notes using the injected noteFetcher.
	notes, err := rp.noteFetcher.GetAllNotes()
	if err != nil {
		log.Printf("ReminderProcessor: Error fetching notes for reminder check: %v", err)
		return // Log the error and exit this check cycle.
	}

	now := time.Now() // Get the current time for comparison.

	// Iterate through each note to check its reminder status.
	for _, n := range notes {
		// Only process notes that have a reminder time explicitly set.
		if n.ReminderTime != nil {
			reminderTime := *n.ReminderTime // Dereference the pointer to get the actual time.Time value.

			// Acquire a lock to safely check and update the `notifiedReminders` map.
			rp.mu.Lock()
			alreadyNotified := rp.notifiedReminders[n.ID] // Check if this reminder has already been notified.
			rp.mu.Unlock()

			// Condition for triggering a notification:
			// 1. The reminder time is in the future (`reminderTime.After(now)`).
			// 2. The reminder time is within the defined `notificationWindow` from now.
			// 3. This specific reminder has NOT been notified yet in this session (`!alreadyNotified`).
			if reminderTime.After(now) && reminderTime.Before(now.Add(rp.notificationWindow)) && !alreadyNotified {
				// Log the reminder notification. In a real application, this would
				// trigger an actual notification (e.g., push notification, email, SMS).
				log.Printf("REMINDER ALERT: Note '%s' (ID: %s) is due at %s",
					n.Title, n.ID, reminderTime.Format(time.RFC3339))

				// Mark this reminder as notified to prevent repeated alerts in subsequent checks.
				rp.mu.Lock()
				rp.notifiedReminders[n.ID] = true
				rp.mu.Unlock()
			} else if reminderTime.Before(now) && alreadyNotified {
				// Optional: If a reminder has passed and was already notified,
				// we could remove it from the `notifiedReminders` map
				// to allow re-notification if the app restarts or the reminder is reset.
				// In a persistent system, a 'notified' flag in the database would be updated.
				rp.mu.Lock()
				delete(rp.notifiedReminders, n.ID) // Clear notification status for past reminders
				rp.mu.Unlock()
			}
		}
	}
}
