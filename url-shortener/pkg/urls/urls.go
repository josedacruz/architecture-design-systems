package urls

import "regexp"

// urlRegex is a basic regex for URL validation.
// In a production system, `net/url.ParseRequestURI` or a more comprehensive library
// would be preferred for robust URL parsing and validation.
var urlRegex = regexp.MustCompile(`^(http|https)://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/\S*)?$`)

// isValidURL performs a basic check to ensure the input string looks like a valid URL.
func IsValidURL(url string) bool {
	return urlRegex.MatchString(url)
}
