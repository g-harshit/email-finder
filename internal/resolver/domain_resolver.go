package resolver

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DomainResolver resolves company names to domains
type DomainResolver struct {
	logger     *zap.Logger
	timeout    time.Duration
	companyMap map[string]string
	mapMutex   sync.RWMutex
}

// wellKnownCompanies is a map of company names (normalized) to their domains
var wellKnownCompanies = map[string]string{
	// Tech Companies
	"google":     "google.com",
	"alphabet":   "google.com",
	"microsoft":  "microsoft.com",
	"apple":      "apple.com",
	"amazon":     "amazon.com",
	"meta":       "meta.com",
	"facebook":   "meta.com",
	"twitter":    "twitter.com",
	"x":          "x.com",
	"linkedin":   "linkedin.com",
	"netflix":    "netflix.com",
	"uber":       "uber.com",
	"airbnb":     "airbnb.com",
	"spotify":    "spotify.com",
	"salesforce": "salesforce.com",
	"oracle":     "oracle.com",
	"ibm":        "ibm.com",
	"intel":      "intel.com",
	"nvidia":     "nvidia.com",
	"adobe":      "adobe.com",
	"paypal":     "paypal.com",
	"stripe":     "stripe.com",
	"shopify":    "shopify.com",
	"zoom":       "zoom.us",
	"slack":      "slack.com",
	"dropbox":    "dropbox.com",
	"tesla":      "tesla.com",
	"spacex":     "spacex.com",

	// Indian Companies
	"zepto":    "zeptonow.com",
	"swiggy":   "swiggy.com",
	"zomato":   "zomato.com",
	"flipkart": "flipkart.com",
	"myntra":   "myntra.com",
	"razorpay": "razorpay.com",
	"paytm":    "paytm.com",
	"phonepe":  "phonepe.com",
	"cred":     "cred.club",
	"byjus":    "byjus.com",
	"infosys":  "infosys.com",
	"tcs":      "tcs.com",
	"tata":     "tata.com",
	"reliance": "reliance.co.in",
	"wipro":    "wipro.com",
	"hcl":      "hcl.com",

	// Financial Services
	"goldman sachs":   "gs.com",
	"morgan stanley":  "morganstanley.com",
	"jpmorgan":        "jpmorgan.com",
	"jpmorgan chase":  "jpmorgan.com",
	"bank of america": "bankofamerica.com",
	"wells fargo":     "wellsfargo.com",
	"citibank":        "citi.com",
	"visa":            "visa.com",
	"mastercard":      "mastercard.com",

	// Consulting & Services
	"mckinsey":               "mckinsey.com",
	"bain":                   "bain.com",
	"boston consulting":      "bcg.com",
	"bcg":                    "bcg.com",
	"deloitte":               "deloitte.com",
	"pwc":                    "pwc.com",
	"pricewaterhousecoopers": "pwc.com",
	"ey":                     "ey.com",
	"ernst young":            "ey.com",
	"kpmg":                   "kpmg.com",
	"accenture":              "accenture.com",

	// Media & Entertainment
	"disney":      "disney.com",
	"warner bros": "warnerbros.com",
	"sony":        "sony.com",
	"nike":        "nike.com",
	"adidas":      "adidas.com",

	// E-commerce & Retail
	"walmart": "walmart.com",
	"target":  "target.com",
	"costco":  "costco.com",
	"ebay":    "ebay.com",
	"etsy":    "etsy.com",

	// Automotive
	"ford":           "ford.com",
	"general motors": "gm.com",
	"gm":             "gm.com",
	"bmw":            "bmw.com",
	"mercedes":       "mercedes-benz.com",
	"volkswagen":     "volkswagen.com",
	"toyota":         "toyota.com",
	"honda":          "honda.com",

	// Food & Beverage
	"starbucks": "starbucks.com",
	"mcdonalds": "mcdonalds.com",
	"coca cola": "coca-cola.com",
	"pepsi":     "pepsico.com",

	// Healthcare & Pharma
	"pfizer":          "pfizer.com",
	"johnson johnson": "jnj.com",
	"jnj":             "jnj.com",
	"novartis":        "novartis.com",
	"roche":           "roche.com",

	// Telecom
	"verizon":  "verizon.com",
	"at&t":     "att.com",
	"att":      "att.com",
	"t-mobile": "t-mobile.com",
	"sprint":   "sprint.com",

	// Energy
	"exxonmobil": "exxonmobil.com",
	"chevron":    "chevron.com",
	"shell":      "shell.com",
	"bp":         "bp.com",
}

// DomainResult represents the result of domain resolution
type DomainResult struct {
	Domain     string   `json:"domain"`
	Resolved   bool     `json:"resolved"`
	Method     string   `json:"method"` // "direct", "pattern", "dns_verified"
	Candidates []string `json:"candidates,omitempty"`
}

// NewDomainResolver creates a new domain resolver
func NewDomainResolver(logger *zap.Logger, timeout time.Duration) *DomainResolver {
	// Initialize company map with well-known companies
	companyMap := make(map[string]string)
	for k, v := range wellKnownCompanies {
		companyMap[k] = v
	}

	return &DomainResolver{
		logger:     logger,
		timeout:    timeout,
		companyMap: companyMap,
	}
}

// AddCompanyDomain adds or updates a company domain mapping
func (r *DomainResolver) AddCompanyDomain(companyName, domain string) {
	r.mapMutex.Lock()
	defer r.mapMutex.Unlock()

	normalized := r.normalizeCompanyName(companyName)
	r.companyMap[normalized] = domain
	r.logger.Debug("added company domain mapping",
		zap.String("company", normalized),
		zap.String("domain", domain),
	)
}

// GetCompanyDomain retrieves a domain for a company if it exists in the map
func (r *DomainResolver) GetCompanyDomain(companyName string) (string, bool) {
	r.mapMutex.RLock()
	defer r.mapMutex.RUnlock()

	normalized := r.normalizeCompanyName(companyName)
	domain, exists := r.companyMap[normalized]
	return domain, exists
}

// ResolveDomain attempts to resolve a company name to a domain
func (r *DomainResolver) ResolveDomain(companyName string) *DomainResult {
	companyName = strings.TrimSpace(strings.ToLower(companyName))

	if companyName == "" {
		return &DomainResult{
			Domain:   "",
			Resolved: false,
			Method:   "none",
		}
	}

	// Check if it's already a domain
	if r.isDomain(companyName) {
		// Verify it has valid DNS records
		if r.verifyDomain(companyName) {
			return &DomainResult{
				Domain:   companyName,
				Resolved: true,
				Method:   "direct",
			}
		}
		// Even if DNS check fails, return it as it might be valid
		return &DomainResult{
			Domain:   companyName,
			Resolved: true,
			Method:   "direct",
		}
	}

	// First, check in-memory company map
	if domain, exists := r.GetCompanyDomain(companyName); exists {
		r.logger.Info("domain resolved from company map",
			zap.String("company", companyName),
			zap.String("domain", domain),
		)
		return &DomainResult{
			Domain:   domain,
			Resolved: true,
			Method:   "company_map",
		}
	}

	// Generate domain candidates
	candidates := r.generateDomainCandidates(companyName)

	// Try to verify candidates via DNS
	for _, candidate := range candidates {
		if r.verifyDomain(candidate) {
			r.logger.Info("domain resolved via DNS",
				zap.String("company", companyName),
				zap.String("domain", candidate),
			)
			return &DomainResult{
				Domain:     candidate,
				Resolved:   true,
				Method:     "dns_verified",
				Candidates: candidates,
			}
		}
	}

	// If no DNS verification, return the most likely candidate
	primaryCandidate := candidates[0]
	r.logger.Info("domain resolved via pattern",
		zap.String("company", companyName),
		zap.String("domain", primaryCandidate),
		zap.Strings("all_candidates", candidates),
	)

	return &DomainResult{
		Domain:     primaryCandidate,
		Resolved:   true,
		Method:     "pattern",
		Candidates: candidates,
	}
}

// isDomain checks if the input looks like a domain
func (r *DomainResolver) isDomain(input string) bool {
	// Simple check: contains at least one dot and no spaces
	if strings.Contains(input, ".") && !strings.Contains(input, " ") {
		parts := strings.Split(input, ".")
		// Should have at least 2 parts (domain.tld)
		if len(parts) >= 2 {
			// Last part should be a valid TLD (2+ characters)
			tld := parts[len(parts)-1]
			return len(tld) >= 2
		}
	}
	return false
}

// generateDomainCandidates generates possible domain names from company name
func (r *DomainResolver) generateDomainCandidates(companyName string) []string {
	candidates := []string{}

	// Clean company name (remove common suffixes, spaces, special chars)
	cleaned := r.cleanCompanyName(companyName)

	// Common TLDs to try
	tlds := []string{"com", "io", "co", "net", "org", "co.uk", "com.au", "ca", "de", "fr"}

	// Generate candidates
	for _, tld := range tlds {
		candidates = append(candidates, fmt.Sprintf("%s.%s", cleaned, tld))
	}

	// Also try with common variations
	// Remove common words like "inc", "llc", "ltd", "corp"
	variations := r.getCompanyVariations(cleaned)
	for _, variation := range variations {
		for _, tld := range tlds {
			candidate := fmt.Sprintf("%s.%s", variation, tld)
			// Avoid duplicates
			exists := false
			for _, existing := range candidates {
				if existing == candidate {
					exists = true
					break
				}
			}
			if !exists {
				candidates = append(candidates, candidate)
			}
		}
	}

	return candidates
}

// normalizeCompanyName normalizes company name for map lookup
func (r *DomainResolver) normalizeCompanyName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Remove common company suffixes
	suffixes := []string{" inc", " llc", " ltd", " corp", " corporation", " limited", " company", " co", " inc.", " llc.", " ltd.", " corp."}
	for _, suffix := range suffixes {
		name = strings.TrimSuffix(name, suffix)
	}

	// Remove leading/trailing spaces
	name = strings.TrimSpace(name)

	return name
}

// cleanCompanyName cleans and normalizes company name for domain generation
func (r *DomainResolver) cleanCompanyName(name string) string {
	// Normalize first
	name = r.normalizeCompanyName(name)

	// Remove special characters (keep only alphanumeric and spaces)
	var cleaned strings.Builder
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == ' ' {
			cleaned.WriteRune(char)
		}
	}

	// Replace spaces with nothing (for domain)
	result := strings.ReplaceAll(cleaned.String(), " ", "")

	// Remove leading/trailing spaces and return
	return strings.TrimSpace(result)
}

// getCompanyVariations returns variations of company name
func (r *DomainResolver) getCompanyVariations(name string) []string {
	variations := []string{name}

	// If name has multiple words, try first word only
	words := strings.Fields(name)
	if len(words) > 1 {
		variations = append(variations, words[0])
		// Try first two words
		if len(words) >= 2 {
			variations = append(variations, words[0]+words[1])
		}
	}

	return variations
}

// verifyDomain checks if a domain has valid DNS records
func (r *DomainResolver) verifyDomain(domain string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	// Try to resolve MX records (most reliable for email domains)
	mxRecords, err := net.DefaultResolver.LookupMX(ctx, domain)
	if err == nil && len(mxRecords) > 0 {
		return true
	}

	// Fallback: try A records
	_, err = net.DefaultResolver.LookupHost(ctx, domain)
	if err == nil {
		return true
	}

	// Fallback: try CNAME
	_, err = net.DefaultResolver.LookupCNAME(ctx, domain)
	if err == nil {
		return true
	}

	return false
}
