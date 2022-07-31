package config

import (
	"errors"
	"os"
)

func exists(path string) bool {
	_, err := os.Stat(path)
	return errors.Is(err, os.ErrNotExist)
}

// StringPtr returns a string pointer
func StringPtr(s string) *string {
	return &s
}

// BoolPtr returns a bool pointer
func BoolPtr(b bool) *bool {
	return &b
}

// Float64Ptr returns a float64 pointer
func Float64Ptr(f float64) *float64 {
	return &f
}
