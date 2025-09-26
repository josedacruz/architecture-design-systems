package util

import (
	"github.com/google/uuid" // External library for generating UUIDs
)

// GenerateUUID generates a new RFC 4122 UUID (version 4).
// UUIDs (Universally Unique Identifiers) are 128-bit numbers used to uniquely
// identify information in computer systems. Version 4 UUIDs are generated
// using random or pseudo-random numbers.
//
// They are suitable for use as primary keys in distributed systems because
// their uniqueness is virtually guaranteed without requiring a central
// coordination service, thus avoiding bottlenecks and single points of failure.
func GenerateUUID() string {
	// uuid.New() generates a new UUID.
	// .String() converts the UUID object into its canonical string representation.
	return uuid.New().String()
}
