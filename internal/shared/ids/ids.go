package ids

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// New generates a new random unique identifier.
// Format: 16 random bytes encoded as a 32-character hex string.
// For production use, replace with UUID v7 for time-sortable IDs.
func New() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("ids.New: entropy source failed: %v", err))
	}
	return hex.EncodeToString(b)
}

// Must panics if the provided ID is empty.
func Must(id string) string {
	if id == "" {
		panic("ids.Must: empty ID")
	}
	return id
}
