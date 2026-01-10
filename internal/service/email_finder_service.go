package service

import (
	"email-finder/internal/generator"
	"email-finder/internal/resolver"
	"email-finder/internal/verifier"

	"go.uber.org/zap"
)

// EmailFinderService handles the core business logic for finding emails
type EmailFinderService struct {
	verifier       verifier.Verifier
	domainResolver *resolver.DomainResolver
	logger         *zap.Logger
	maxPatterns    int
}

// NewEmailFinderService creates a new email finder service
func NewEmailFinderService(v verifier.Verifier, dr *resolver.DomainResolver, logger *zap.Logger, maxPatterns int) *EmailFinderService {
	return &EmailFinderService{
		verifier:       v,
		domainResolver: dr,
		logger:         logger,
		maxPatterns:    maxPatterns,
	}
}

// FindEmailRequest represents the input for finding emails
type FindEmailRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Company   string `json:"company" binding:"required"`
}

// EmailResult represents a found email with verification details
type EmailResult struct {
	Email         string `json:"email"`
	Pattern       string `json:"pattern"`
	IsReachable   string `json:"is_reachable"`
	IsValid       bool   `json:"is_valid"`
	IsDeliverable bool   `json:"is_deliverable"`
	Confidence    string `json:"confidence"` // high, medium, low
}

// FindEmailResponse represents the response from finding emails
type FindEmailResponse struct {
	FoundEmails    []EmailResult    `json:"found_emails"`
	TotalChecked   int              `json:"total_checked"`
	TotalFound     int              `json:"total_found"`
	Domain         string           `json:"domain"`
	DomainResolved bool             `json:"domain_resolved"`
	Request        FindEmailRequest `json:"request"`
}

// FindEmails finds and verifies emails based on the input
func (s *EmailFinderService) FindEmails(req FindEmailRequest) (*FindEmailResponse, error) {
	s.logger.Info("finding emails",
		zap.String("first_name", req.FirstName),
		zap.String("last_name", req.LastName),
		zap.String("company", req.Company),
	)

	// Resolve domain from company name
	domainResult := s.domainResolver.ResolveDomain(req.Company)
	domain := domainResult.Domain

	if !domainResult.Resolved || domain == "" {
		s.logger.Warn("failed to resolve domain",
			zap.String("company", req.Company),
		)
		return &FindEmailResponse{
			FoundEmails:    []EmailResult{},
			TotalChecked:   0,
			TotalFound:     0,
			Domain:         "",
			DomainResolved: false,
			Request:        req,
		}, nil
	}

	s.logger.Info("domain resolved",
		zap.String("company", req.Company),
		zap.String("domain", domain),
		zap.String("method", domainResult.Method),
	)

	// Generate email patterns using resolved domain
	patterns := generator.GenerateEmailPatterns(req.FirstName, req.LastName, domain)

	// Patterns are already generated in priority order (base patterns first, then numbered)
	// This ensures common patterns are verified first, improving perceived latency

	// Limit the number of patterns if configured
	if s.maxPatterns > 0 && len(patterns) > s.maxPatterns {
		patterns = patterns[:s.maxPatterns]
	}

	if len(patterns) == 0 {
		return &FindEmailResponse{
			FoundEmails:    []EmailResult{},
			TotalChecked:   0,
			TotalFound:     0,
			Domain:         domain,
			DomainResolved: true,
			Request:        req,
		}, nil
	}

	// Extract emails for verification
	emails := make([]string, 0, len(patterns))
	emailToPattern := make(map[string]string)
	for _, pattern := range patterns {
		emails = append(emails, pattern.Email)
		emailToPattern[pattern.Email] = pattern.Pattern
	}

	// Verify emails
	verificationResults, err := s.verifier.VerifyEmailsBatch(emails)
	if err != nil {
		s.logger.Error("failed to verify emails", zap.Error(err))
		return nil, err
	}

	// Process results and filter valid emails
	// Only return emails that are verified and deliverable
	foundEmails := make([]EmailResult, 0)
	for _, result := range verificationResults {
		// Only include emails that are verified (not unknown) and deliverable
		if result.IsReachable != "unknown" && (result.IsReachable == "safe" || (result.IsReachable == "risky" && result.IsDeliverable)) {
			confidence := s.calculateConfidence(result)
			foundEmails = append(foundEmails, EmailResult{
				Email:         result.Email,
				Pattern:       emailToPattern[result.Email],
				IsReachable:   result.IsReachable,
				IsValid:       result.IsValid,
				IsDeliverable: result.IsDeliverable,
				Confidence:    confidence,
			})
		}
	}

	// Sort by confidence (high to low)
	foundEmails = s.sortByConfidence(foundEmails)

	s.logger.Info("email search completed",
		zap.Int("total_checked", len(patterns)),
		zap.Int("total_found", len(foundEmails)),
	)

	return &FindEmailResponse{
		FoundEmails:    foundEmails,
		TotalChecked:   len(patterns),
		TotalFound:     len(foundEmails),
		Domain:         domain,
		DomainResolved: true,
		Request:        req,
	}, nil
}

// calculateConfidence determines the confidence level for an email
func (s *EmailFinderService) calculateConfidence(result *verifier.VerificationResult) string {
	if result.IsReachable == "safe" && result.IsDeliverable {
		return "high"
	}
	if result.IsReachable == "risky" && result.IsDeliverable {
		return "medium"
	}
	if result.IsValid {
		return "low"
	}
	return "low"
}

// sortByConfidence sorts emails by confidence level
func (s *EmailFinderService) sortByConfidence(emails []EmailResult) []EmailResult {
	// Simple sort: high, medium, low
	high := make([]EmailResult, 0)
	medium := make([]EmailResult, 0)
	low := make([]EmailResult, 0)

	for _, email := range emails {
		switch email.Confidence {
		case "high":
			high = append(high, email)
		case "medium":
			medium = append(medium, email)
		default:
			low = append(low, email)
		}
	}

	result := make([]EmailResult, 0, len(emails))
	result = append(result, high...)
	result = append(result, medium...)
	result = append(result, low...)

	return result
}
