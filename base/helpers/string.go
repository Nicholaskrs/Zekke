package helpers

import (
	"regexp"
	"strconv"
)

// IsAlphanumeric used to check if string is alphanumeric
// return true if alphanumeric
func IsAlphanumeric(str string) bool {
	// Regular expression pattern for alphanumeric characters
	pattern := "^[a-zA-Z0-9]+$"

	// Compile the regular expression
	regExp := regexp.MustCompile(pattern)

	// Use the MatchString method to check if the string matches the pattern
	return regExp.MatchString(str)
}

// IsAlphanumericSpecialChar used to check if string is alphanumeric and special character (include spaces)
// return true if alphanumeric (include spaces)
func IsAlphanumericSpecialChar(str string) bool {
	// Regular expression pattern for alphanumeric characters
	pattern := `^[a-zA-Z0-9\s.,?!@#$%^&*()_-]+$`

	// Compile the regular expression
	regExp := regexp.MustCompile(pattern)

	// Use the MatchString method to check if the string matches the pattern
	return regExp.MatchString(str)
}

// IsNumeric used to check if a string represents a numeric value
// return true if string is numeric
func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
