package generator

import (
	"testing"
)

func TestGenerateEmailPatterns(t *testing.T) {
	tests := []struct {
		name     string
		firstName string
		lastName  string
		domain    string
		wantMin   int
	}{
		{
			name:      "normal case",
			firstName: "john",
			lastName:  "doe",
			domain:    "example.com",
			wantMin:   10,
		},
		{
			name:      "single character first name",
			firstName: "j",
			lastName:  "doe",
			domain:    "example.com",
			wantMin:   5,
		},
		{
			name:      "empty inputs",
			firstName: "",
			lastName:  "",
			domain:    "",
			wantMin:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := GenerateEmailPatterns(tt.firstName, tt.lastName, tt.domain)
			if len(patterns) < tt.wantMin {
				t.Errorf("GenerateEmailPatterns() generated %d patterns, want at least %d", len(patterns), tt.wantMin)
			}

			// Check that all patterns are valid
			for _, pattern := range patterns {
				if pattern.Email == "" {
					t.Errorf("GenerateEmailPatterns() generated empty email pattern")
				}
				if !isValidEmailFormat(pattern.Email) {
					t.Errorf("GenerateEmailPatterns() generated invalid email: %s", pattern.Email)
				}
			}
		})
	}
}

func TestIsValidEmailFormat(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid email", "john.doe@example.com", true},
		{"valid email with numbers", "john123@example.com", true},
		{"invalid no @", "johndoeexample.com", false},
		{"invalid no domain", "john@", false},
		{"invalid no local", "@example.com", false},
		{"empty", "", false},
		{"valid with underscore", "john_doe@example.com", true},
		{"valid with dash", "john-doe@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidEmailFormat(tt.email); got != tt.want {
				t.Errorf("isValidEmailFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
