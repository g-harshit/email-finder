package resolver

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestDomainResolver_ResolveDomain(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	resolver := NewDomainResolver(logger, 5*time.Second)

	tests := []struct {
		name     string
		company  string
		wantResolved bool
		wantDomain   string
		wantMethod   string
	}{
		{
			name:     "already a domain",
			company:  "example.com",
			wantResolved: true,
			wantDomain: "example.com",
			wantMethod: "direct",
		},
		{
			name:     "company name from map",
			company:  "Google",
			wantResolved: true,
			wantDomain: "google.com",
			wantMethod: "company_map",
		},
		{
			name:     "company with suffix from map",
			company:  "Microsoft Inc",
			wantResolved: true,
			wantDomain: "microsoft.com",
			wantMethod: "company_map",
		},
		{
			name:     "zepto from map",
			company:  "Zepto",
			wantResolved: true,
			wantDomain: "zeptonow.com",
			wantMethod: "company_map",
		},
		{
			name:     "empty string",
			company:  "",
			wantResolved: false,
			wantDomain: "",
			wantMethod: "none",
		},
		{
			name:     "company with spaces (not in map)",
			company:  "Acme Corporation",
			wantResolved: true,
			wantDomain: "", // Will be resolved via DNS/pattern
			wantMethod: "", // Will be dns_verified or pattern
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolver.ResolveDomain(tt.company)
			if result.Resolved != tt.wantResolved {
				t.Errorf("ResolveDomain() Resolved = %v, want %v", result.Resolved, tt.wantResolved)
			}
			if result.Resolved && result.Domain == "" && tt.wantDomain != "" {
				t.Errorf("ResolveDomain() Resolved = true but Domain is empty")
			}
			if tt.wantDomain != "" && result.Domain != tt.wantDomain {
				t.Errorf("ResolveDomain() Domain = %v, want %v", result.Domain, tt.wantDomain)
			}
			if tt.wantMethod != "" && result.Method != tt.wantMethod {
				t.Errorf("ResolveDomain() Method = %v, want %v", result.Method, tt.wantMethod)
			}
		})
	}
}

func TestIsDomain(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	resolver := NewDomainResolver(logger, 5*time.Second)

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid domain", "example.com", true},
		{"valid domain with subdomain", "mail.example.com", true},
		{"not a domain", "example", false},
		{"not a domain with space", "example com", false},
		{"empty", "", false},
		{"domain with short tld", "example.c", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolver.isDomain(tt.input); got != tt.want {
				t.Errorf("isDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCleanCompanyName(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	resolver := NewDomainResolver(logger, 5*time.Second)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"with inc", "acme inc", "acme"},
		{"with llc", "acme llc", "acme"},
		{"with ltd", "acme ltd", "acme"},
		{"with corp", "acme corp", "acme"},
		{"with special chars", "acme-corp!", "acmecorp"},
		{"multiple words", "acme corporation", "acme"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolver.cleanCompanyName(tt.input)
			if got != tt.want {
				t.Errorf("cleanCompanyName() = %v, want %v", got, tt.want)
			}
		})
	}
}
