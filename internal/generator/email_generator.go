package generator

import (
	"fmt"
	"strings"
	"unicode"
)

// EmailPattern represents a generated email pattern
type EmailPattern struct {
	Email   string
	Pattern string
}

// GenerateEmailPatterns generates all possible email patterns based on first name, last name, and domain
func GenerateEmailPatterns(firstName, lastName, domain string) []EmailPattern {
	patterns := []EmailPattern{}

	// Normalize inputs
	firstName = strings.TrimSpace(strings.ToLower(firstName))
	lastName = strings.TrimSpace(strings.ToLower(lastName))
	domain = strings.TrimSpace(strings.ToLower(domain))

	if firstName == "" || lastName == "" || domain == "" {
		return patterns
	}

	// Get first letter of first name
	firstInitial := ""
	if len(firstName) > 0 {
		firstInitial = string(firstName[0])
	}

	// Get first letter of last name
	lastInitial := ""
	if len(lastName) > 0 {
		lastInitial = string(lastName[0])
	}

	// Common email patterns
	patternList := []struct {
		email   string
		pattern string
	}{
		{firstName + "." + lastName + "@" + domain, "firstname.lastname"},
		{firstName + lastName + "@" + domain, "firstnamelastname"},
		{firstInitial + "." + lastName + "@" + domain, "f.lastname"},
		{firstInitial + lastName + "@" + domain, "flastname"},
		{firstName + "." + lastInitial + "@" + domain, "firstname.l"},
		{firstName + lastInitial + "@" + domain, "firstnamel"},
		{firstName + "@" + domain, "firstname"},
		{lastName + "@" + domain, "lastname"},
		{lastName + "." + firstName + "@" + domain, "lastname.firstname"},
		{lastName + firstName + "@" + domain, "lastnamefirstname"},
		{lastInitial + "." + firstName + "@" + domain, "l.firstname"},
		{lastInitial + firstName + "@" + domain, "lfirstname"},
		{firstInitial + "_" + lastName + "@" + domain, "f_lastname"},
		{firstName + "_" + lastName + "@" + domain, "firstname_lastname"},
		{lastName + "_" + firstName + "@" + domain, "lastname_firstname"},
		{firstInitial + lastInitial + "@" + domain, "fl"},
		{firstName + "." + firstInitial + "." + lastName + "@" + domain, "firstname.f.lastname"},
		{lastName + "." + firstInitial + "@" + domain, "lastname.f"},
		{firstInitial + "." + firstName + "." + lastName + "@" + domain, "f.firstname.lastname"},
		{firstName + "-" + lastName + "@" + domain, "firstname-lastname"},
	}

	// Add patterns with numbers (0-9) for common variations
	for i := 0; i <= 9; i++ {
		num := string(rune('0' + i))
		patternList = append(patternList,
			struct {
				email   string
				pattern string
			}{firstName + "." + lastName + num + "@" + domain, "firstname.lastname" + num},
			struct {
				email   string
				pattern string
			}{firstName + lastName + num + "@" + domain, "firstnamelastname" + num},
			struct {
				email   string
				pattern string
			}{firstInitial + "." + lastName + num + "@" + domain, "f.lastname" + num},
		)
	}

	// Add patterns with numbers 1-50 for firstname.lastname variations
	for i := 1; i <= 50; i++ {
		num := fmt.Sprintf("%d", i)
		patternList = append(patternList,
			struct {
				email   string
				pattern string
			}{firstName + "." + lastName + num + "@" + domain, "firstname.lastname" + num},
			struct {
				email   string
				pattern string
			}{firstName + lastName + num + "@" + domain, "firstnamelastname" + num},
			struct {
				email   string
				pattern string
			}{firstInitial + "." + lastName + num + "@" + domain, "f.lastname" + num},
		)
	}

	// Convert to EmailPattern and remove duplicates
	seen := make(map[string]bool)
	for _, p := range patternList {
		if !seen[p.email] && isValidEmailFormat(p.email) {
			patterns = append(patterns, EmailPattern{
				Email:   p.email,
				Pattern: p.pattern,
			})
			seen[p.email] = true
		}
	}

	return patterns
}

// isValidEmailFormat performs basic email format validation
func isValidEmailFormat(email string) bool {
	if len(email) == 0 || len(email) > 254 {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]

	if len(local) == 0 || len(local) > 64 {
		return false
	}

	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// Check local part (before @)
	for _, char := range local {
		if !isValidEmailChar(char) {
			return false
		}
	}

	// Check domain part
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 {
		return false
	}

	for _, part := range domainParts {
		if len(part) == 0 {
			return false
		}
		for _, char := range part {
			if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '-' {
				return false
			}
		}
	}

	return true
}

// isValidEmailChar checks if a character is valid in email local part
func isValidEmailChar(char rune) bool {
	return unicode.IsLetter(char) ||
		unicode.IsDigit(char) ||
		char == '.' ||
		char == '_' ||
		char == '-' ||
		char == '+'
}
