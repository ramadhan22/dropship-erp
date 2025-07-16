package service

import (
	"strings"
	"unicode"
)

func sanitizeID(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

// ptrString helper
func ptrString(s string) *string { return &s }

// stringPtr helper (alias for ptrString)
func stringPtr(s string) *string { return &s }
