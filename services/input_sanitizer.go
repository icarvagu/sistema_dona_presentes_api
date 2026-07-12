package services

import "regexp"

var reNonPrintable = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F]`)

// SanitizeString removes non-printable control characters and truncates the
// string to a maximum of 1000 runes. Used for general text input sanitization.
func SanitizeString(s string) string {
	cleaned := reNonPrintable.ReplaceAllString(s, "")
	runes := []rune(cleaned)
	if len(runes) > 1000 {
		cleaned = string(runes[:1000])
	}
	return cleaned
}

// SanitizeUsername removes non-printable control characters and truncates to
// a maximum of 255 bytes. Used for username input sanitization.
func SanitizeUsername(s string) string {
	cleaned := reNonPrintable.ReplaceAllString(s, "")
	if len(cleaned) > 255 {
		cleaned = cleaned[:255]
	}
	return cleaned
}

// SanitizeCPF strips all non-digit characters and truncates to 11 digits maximum.
func SanitizeCPF(s string) string {
	digitsOnly := regexp.MustCompile(`[^\d]`).ReplaceAllString(s, "")
	if len(digitsOnly) > 11 {
		digitsOnly = digitsOnly[:11]
	}
	return digitsOnly
}

// SanitizePassword truncates the password to a maximum of 128 characters.
func SanitizePassword(s string) string {
	if len(s) > 128 {
		s = s[:128]
	}
	return s
}
