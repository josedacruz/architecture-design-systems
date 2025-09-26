package constants

import (
	"time"
)

const (
	DefaultPort                = "8080"
	ReminderCheckInterval      = 10 * time.Second // How often the processor checks for reminders
	ReminderNotificationWindow = 5 * time.Minute  // Notify if reminder is within this window
	SqliteDBPath               = "./notes.db"     // Path for the SQLite database file
)
