package verifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"go.uber.org/zap"
)

// VerificationResult represents the result of email verification
type VerificationResult struct {
	Email         string                 `json:"email"`
	IsReachable   string                 `json:"is_reachable"` // safe, risky, invalid, unknown
	IsValid       bool                   `json:"is_valid"`
	IsDeliverable bool                   `json:"is_deliverable"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// Verifier interface for email verification
type Verifier interface {
	VerifyEmail(email string) (*VerificationResult, error)
	VerifyEmailsBatch(emails []string) ([]*VerificationResult, error)
}

// HTTPVerifier uses the check-if-email-exists HTTP API
type HTTPVerifier struct {
	apiURL      string
	apiEndpoint string
	client      *http.Client
	logger      *zap.Logger
	timeout     time.Duration
	concurrency int
}

// CLIVerifier uses the check-if-email-exists CLI binary
type CLIVerifier struct {
	cliPath     string
	logger      *zap.Logger
	timeout     time.Duration
	concurrency int
}

// NewHTTPVerifier creates a new HTTP-based verifier
func NewHTTPVerifier(apiURL, apiEndpoint string, timeout time.Duration, concurrency int, logger *zap.Logger) *HTTPVerifier {
	if concurrency <= 0 {
		concurrency = 10 // Default concurrency
	}
	return &HTTPVerifier{
		apiURL:      apiURL,
		apiEndpoint: apiEndpoint,
		client: &http.Client{
			Timeout: timeout,
		},
		logger:      logger,
		timeout:     timeout,
		concurrency: concurrency,
	}
}

// NewCLIVerifier creates a new CLI-based verifier
func NewCLIVerifier(cliPath string, timeout time.Duration, concurrency int, logger *zap.Logger) *CLIVerifier {
	if concurrency <= 0 {
		concurrency = 10 // Default concurrency
	}
	return &CLIVerifier{
		cliPath:     cliPath,
		logger:      logger,
		timeout:     timeout,
		concurrency: concurrency,
	}
}

// VerifyEmail verifies a single email using HTTP API
func (v *HTTPVerifier) VerifyEmail(email string) (*VerificationResult, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"to_email": email,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	url := fmt.Sprintf("%s%s", v.apiURL, v.apiEndpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		v.logger.Warn("verification API returned non-200 status",
			zap.String("email", email),
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return &VerificationResult{
			Email:         email,
			IsReachable:   "unknown",
			IsValid:       false,
			IsDeliverable: false,
		}, nil
	}

	// Parse response
	var apiResponse struct {
		Input       string `json:"input"`
		IsReachable string `json:"is_reachable"`
		SMTP        struct {
			IsDeliverable bool `json:"is_deliverable"`
		} `json:"smtp"`
		Syntax struct {
			IsValidSyntax bool `json:"is_valid_syntax"`
		} `json:"syntax"`
		MX struct {
			AcceptsMail bool `json:"accepts_mail"`
		} `json:"mx"`
	}

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		// Try to parse as array (some APIs return array)
		var responses []struct {
			Input       string `json:"input"`
			IsReachable string `json:"is_reachable"`
			SMTP        struct {
				IsDeliverable bool `json:"is_deliverable"`
			} `json:"smtp"`
			Syntax struct {
				IsValidSyntax bool `json:"is_valid_syntax"`
			} `json:"syntax"`
			MX struct {
				AcceptsMail bool `json:"accepts_mail"`
			} `json:"mx"`
		}

		if err2 := json.Unmarshal(body, &responses); err2 != nil || len(responses) == 0 {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		apiResponse = struct {
			Input       string `json:"input"`
			IsReachable string `json:"is_reachable"`
			SMTP        struct {
				IsDeliverable bool `json:"is_deliverable"`
			} `json:"smtp"`
			Syntax struct {
				IsValidSyntax bool `json:"is_valid_syntax"`
			} `json:"syntax"`
			MX struct {
				AcceptsMail bool `json:"accepts_mail"`
			} `json:"mx"`
		}{
			Input:       responses[0].Input,
			IsReachable: responses[0].IsReachable,
			SMTP:        responses[0].SMTP,
			Syntax:      responses[0].Syntax,
			MX:          responses[0].MX,
		}
	}

	result := &VerificationResult{
		Email:         apiResponse.Input,
		IsReachable:   apiResponse.IsReachable,
		IsValid:       apiResponse.Syntax.IsValidSyntax && apiResponse.MX.AcceptsMail,
		IsDeliverable: apiResponse.SMTP.IsDeliverable,
		Details: map[string]interface{}{
			"syntax_valid": apiResponse.Syntax.IsValidSyntax,
			"mx_accepts":   apiResponse.MX.AcceptsMail,
		},
	}

	return result, nil
}

// VerifyEmailsBatch verifies multiple emails in parallel
func (v *HTTPVerifier) VerifyEmailsBatch(emails []string) ([]*VerificationResult, error) {
	if len(emails) == 0 {
		return []*VerificationResult{}, nil
	}

	results := make([]*VerificationResult, len(emails))
	semaphore := make(chan struct{}, v.concurrency)
	var wg sync.WaitGroup

	for i, email := range emails {
		wg.Add(1)
		go func(idx int, emailAddr string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result, err := v.VerifyEmail(emailAddr)
			if err != nil {
				v.logger.Error("failed to verify email",
					zap.String("email", emailAddr),
					zap.Error(err),
				)
				result = &VerificationResult{
					Email:         emailAddr,
					IsReachable:   "unknown",
					IsValid:       false,
					IsDeliverable: false,
				}
			}
			results[idx] = result
		}(i, email)
	}

	wg.Wait()
	return results, nil
}

// VerifyEmail verifies a single email using CLI
func (v *CLIVerifier) VerifyEmail(email string) (*VerificationResult, error) {
	// Use a shorter timeout per email to prevent hanging
	emailTimeout := 10 * time.Second
	if v.timeout < emailTimeout {
		emailTimeout = v.timeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), emailTimeout)
	defer cancel()

	// Use exec.CommandContext for timeout support
	cmd := exec.CommandContext(ctx, v.cliPath, email)
	output, err := cmd.Output()
	if err != nil {
		// Check if it's a timeout error
		if ctx.Err() == context.DeadlineExceeded {
			return &VerificationResult{
				Email:         email,
				IsReachable:   "unknown",
				IsValid:       false,
				IsDeliverable: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to execute CLI: %w", err)
	}

	// Parse JSON output from CLI
	var apiResponse struct {
		Input       string `json:"input"`
		IsReachable string `json:"is_reachable"`
		SMTP        struct {
			IsDeliverable bool `json:"is_deliverable"`
		} `json:"smtp"`
		Syntax struct {
			IsValidSyntax bool `json:"is_valid_syntax"`
		} `json:"syntax"`
		MX struct {
			AcceptsMail bool `json:"accepts_mail"`
		} `json:"mx"`
	}

	if err := json.Unmarshal(output, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse CLI output: %w", err)
	}

	result := &VerificationResult{
		Email:         apiResponse.Input,
		IsReachable:   apiResponse.IsReachable,
		IsValid:       apiResponse.Syntax.IsValidSyntax && apiResponse.MX.AcceptsMail,
		IsDeliverable: apiResponse.SMTP.IsDeliverable,
		Details: map[string]interface{}{
			"syntax_valid": apiResponse.Syntax.IsValidSyntax,
			"mx_accepts":   apiResponse.MX.AcceptsMail,
		},
	}

	return result, nil
}

// VerifyEmailsBatch verifies multiple emails in parallel using CLI
// Optimized: Inlines VerifyEmail to avoid function call overhead and uses per-email timeout
func (v *CLIVerifier) VerifyEmailsBatch(emails []string) ([]*VerificationResult, error) {
	if len(emails) == 0 {
		return []*VerificationResult{}, nil
	}

	results := make([]*VerificationResult, len(emails))
	semaphore := make(chan struct{}, v.concurrency)
	var wg sync.WaitGroup

	// Use a shorter timeout per email to prevent slow verifications from blocking others
	// Aggressively reduced to 3 seconds - most verifications complete in 1-2 seconds
	// Slow verifications will timeout and be marked as unknown, allowing faster overall completion
	perEmailTimeout := 3 * time.Second
	if v.timeout < perEmailTimeout {
		perEmailTimeout = v.timeout
	}

	for i, email := range emails {
		wg.Add(1)
		go func(idx int, emailAddr string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Inline verification to avoid function call overhead
			ctx, cancel := context.WithTimeout(context.Background(), perEmailTimeout)
			defer cancel()

			cmd := exec.CommandContext(ctx, v.cliPath, emailAddr)
			output, err := cmd.Output()

			var result *VerificationResult
			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					// Timeout - mark as unknown
					result = &VerificationResult{
						Email:         emailAddr,
						IsReachable:   "unknown",
						IsValid:       false,
						IsDeliverable: false,
					}
				} else {
					v.logger.Error("failed to verify email",
						zap.String("email", emailAddr),
						zap.Error(err),
					)
					result = &VerificationResult{
						Email:         emailAddr,
						IsReachable:   "unknown",
						IsValid:       false,
						IsDeliverable: false,
					}
				}
			} else {
				// Parse JSON output from CLI
				var apiResponse struct {
					Input       string `json:"input"`
					IsReachable string `json:"is_reachable"`
					SMTP        struct {
						IsDeliverable bool `json:"is_deliverable"`
					} `json:"smtp"`
					Syntax struct {
						IsValidSyntax bool `json:"is_valid_syntax"`
					} `json:"syntax"`
					MX struct {
						AcceptsMail bool `json:"accepts_mail"`
					} `json:"mx"`
				}

				if err := json.Unmarshal(output, &apiResponse); err != nil {
					v.logger.Error("failed to parse CLI output",
						zap.String("email", emailAddr),
						zap.Error(err),
					)
					result = &VerificationResult{
						Email:         emailAddr,
						IsReachable:   "unknown",
						IsValid:       false,
						IsDeliverable: false,
					}
				} else {
					result = &VerificationResult{
						Email:         apiResponse.Input,
						IsReachable:   apiResponse.IsReachable,
						IsValid:       apiResponse.Syntax.IsValidSyntax && apiResponse.MX.AcceptsMail,
						IsDeliverable: apiResponse.SMTP.IsDeliverable,
						Details: map[string]interface{}{
							"syntax_valid": apiResponse.Syntax.IsValidSyntax,
							"mx_accepts":   apiResponse.MX.AcceptsMail,
						},
					}
				}
			}
			results[idx] = result
		}(i, email)
	}

	wg.Wait()
	return results, nil
}
