// main.go (Go Server)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// UserSecret stores the secret key for a user
type UserSecret struct {
	Secret string `json:"secret"`
}

// RegisterRequest represents the request body for /register
type RegisterRequest struct {
	Username string `json:"username"`
}

// RegisterResponse represents the response body for /register
type RegisterResponse struct {
	Message  string `json:"message"`
	Username string `json:"username"`
	Secret   string `json:"secret"`
	TOTPUri  string `json:"totp_uri"`
}

// VerifyRequest represents the request body for /verify
type VerifyRequest struct {
	Username string `json:"username"`
	TOTPCode string `json:"totp_code"`
}

// VerifyResponse represents the response body for /verify
type VerifyResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// In-memory storage for user secrets (for demonstration purposes only)
// In a real application, this would be stored securely in a database.
var userSecrets = make(map[string]string)
var mu sync.Mutex // Mutex to protect access to userSecrets map

func main() {
	// Set up HTTP routes
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/verify", verifyHandler)

	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Starting Go TOTP server on %s\n", addr)
	// Start the HTTP server
	log.Fatal(http.ListenAndServe(addr, nil))
}

// rootHandler provides a basic server check
func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Go TOTP Server is running!")
}

// registerHandler handles user registration for 2FA
func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins for local demo
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Generate a new random secret key for the user
	// otp.GenerateOptions provides options for key generation, including issuer and account name
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "GoTOTPDemo",
		AccountName: req.Username,
		SecretSize:  20, // 20 bytes for a 160-bit secret
		Digits:      otp.DigitsSix,
		Period:      30, // 30 seconds period
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		log.Printf("Error generating TOTP key: %v", err)
		http.Error(w, "Failed to generate secret", http.StatusInternalServerError)
		return
	}

	userSecrets[req.Username] = key.Secret() // Store the Base32 encoded secret

	log.Printf("Registered %s with secret: %s\n", req.Username, key.Secret())
	log.Printf("TOTP URI for %s: %s\n", req.Username, key.URL())

	resp := RegisterResponse{
		Message:  "2FA registration successful",
		Username: req.Username,
		Secret:   key.Secret(), // Return the secret for the Go client's mimicked authenticator
		TOTPUri:  key.URL(),    // Return the provisioning URI (for QR code generation if needed elsewhere)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// verifyHandler handles TOTP code verification
func verifyHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins for local demo
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.TOTPCode == "" {
		http.Error(w, "Username and TOTP code are required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	secret, ok := userSecrets[req.Username]
	mu.Unlock()

	if !ok {
		resp := VerifyResponse{
			Message: "User not registered for 2FA",
			Success: false,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Verify the TOTP code.
	// totp.Validate takes the code, the secret, and an optional validation options struct.
	// With the default options, it checks the current window and allows for a small time skew.
	// The Period and Digits are derived from the secret's provisioning URI or can be specified.
	// The library handles the time window (default 1 step before/after) automatically.
	isValid := totp.Validate(req.TOTPCode, secret)
	// it has a skew that allows the previous for 30s (check the function at the end)

	resp := VerifyResponse{}
	if isValid {
		log.Printf("TOTP verification successful for %s\n", req.Username)
		resp.Message = "TOTP verification successful"
		resp.Success = true
	} else {
		log.Printf("TOTP verification failed for %s with code %s\n", req.Username, req.TOTPCode)
		resp.Message = "Invalid TOTP code"
		resp.Success = false
		w.WriteHeader(http.StatusUnauthorized)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// if you want to be strict on the 30s replace the validation in the code to a call to this function
func verifyTOTP(code, secret string) (bool, error) {
	return totp.ValidateCustom(
		code,
		secret,
		time.Now(), // use server time
		totp.ValidateOpts{
			Period:    30,                // 30s step
			Skew:      0,                 // no previous/next step allowed
			Digits:    otp.DigitsSix,     // 6 digits
			Algorithm: otp.AlgorithmSHA1, // TOTP default
		},
	)
}
