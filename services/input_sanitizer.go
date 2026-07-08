package services

import "regexp"

var reNonPrintable = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F]`)

func SanitizeString(s string) string {
	cleaned := reNonPrintable.ReplaceAllString(s, "")
	if len(cleaned) > 1000 {
		cleaned = cleaned[:1000]
	}
	return cleaned
}

func SanitizeUsername(s string) string {
	cleaned := reNonPrintable.ReplaceAllString(s, "")
	if len(cleaned) > 255 {
		cleaned = cleaned[:255]
	}
	return cleaned
}

func SanitizeCPF(s string) string {
	digitsOnly := regexp.MustCompile(`[^\d]`).ReplaceAllString(s, "")
	if len(digitsOnly) > 11 {
		digitsOnly = digitsOnly[:11]
	}
	return digitsOnly
}

func SanitizePassword(s string) string {
	if len(s) > 128 {
		s = s[:128]
	}
	return s
}
