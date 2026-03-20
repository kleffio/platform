// Package config provides helpers for reading configuration from environment
// variables with typed defaults.
package config

import (
	"os"
	"strconv"
)

// String reads an environment variable, returning the fallback if unset or empty.
func String(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Int reads an environment variable as an integer, returning the fallback on
// parse failure or if the variable is unset.
func Int(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// Bool reads an environment variable as a boolean.
// Accepted truthy values: "1", "true", "yes". Everything else is false.
func Bool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes"
}
