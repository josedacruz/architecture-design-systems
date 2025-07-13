package model

// ShortenRequest defines the structure for the request body when shortening a URL.
type ShortenRequest struct {
	LongURL string `json:"long_url"` // The original long URL to be shortened
}

// ShortenResponse defines the structure for the response body after shortening a URL.
type ShortenResponse struct {
	ShortURL string `json:"short_url"` // The newly generated short URL
}

// ErrorResponse defines a generic error response structure for API errors.
type ErrorResponse struct {
	Message string `json:"message"` // A descriptive error message
}
