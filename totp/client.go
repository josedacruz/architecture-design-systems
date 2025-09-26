// main.go (Go Client)
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	totpPeriod = 30 // Standard TOTP period in seconds
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <base32_secret_key>")
		fmt.Println("Example: go run main.go JBSWY3DPEHPK3PXP")
		fmt.Println("\nTo get a secret key, run the Go server and hit its /register endpoint.")
		os.Exit(1)
	}

	secretKey := os.Args[1]
	fmt.Printf("Starting TOTP authenticator with secret: %s\n", secretKey)
	fmt.Printf("Generating new code every %d seconds...\n", totpPeriod)

	opts := totp.ValidateOpts{
		Period:    uint(totpPeriod),
		Skew:      0,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}

	for {
		now := time.Now()

		// Generate code for the current time
		code, err := totp.GenerateCodeCustom(secretKey, now, opts)
		if err != nil {
			log.Fatalf("Error generating TOTP code: %v", err)
		}

		// Seconds remaining in the current period
		secs := now.Unix()
		timeRemaining := totpPeriod - int(secs%int64(totpPeriod))

		fmt.Printf("Current TOTP Code: %s | Time Remaining: %2d seconds\n", code, timeRemaining)
		time.Sleep(1 * time.Second)
	}
}
