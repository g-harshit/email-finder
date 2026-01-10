package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	Server                  ServerConfig
	EmailVerification       EmailVerificationConfig
	Logging                 LoggingConfig
	RateLimit               int
	VerificationTimeout     time.Duration
	MaxEmailPatterns        int
	VerificationConcurrency int
}

type ServerConfig struct {
	Port string
	Host string
}

type EmailVerificationConfig struct {
	APIURL      string
	APIEndpoint string
	CLIPath     string
	UseCLI      bool
}

type LoggingConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if it doesn't)
	_ = godotenv.Load()

	port := getEnv("SERVER_PORT", "8080")
	host := getEnv("SERVER_HOST", "0.0.0.0")

	apiURL := getEnv("EMAIL_VERIFICATION_API_URL", "http://localhost:8081")
	apiEndpoint := getEnv("EMAIL_VERIFICATION_API_ENDPOINT", "/v0/check_email")
	cliPath := getEnv("EMAIL_VERIFICATION_CLI_PATH", "")
	useCLI := cliPath != ""

	logLevel := getEnv("LOG_LEVEL", "info")
	logFormat := getEnv("LOG_FORMAT", "json")

	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT", "60"))
	timeoutSeconds, _ := strconv.Atoi(getEnv("VERIFICATION_TIMEOUT", "30"))
	maxPatterns, _ := strconv.Atoi(getEnv("MAX_EMAIL_PATTERNS", "200")) // Increased default for numbered patterns
	verificationConcurrency, _ := strconv.Atoi(getEnv("VERIFICATION_CONCURRENCY", "100"))

	config := &Config{
		Server: ServerConfig{
			Port: port,
			Host: host,
		},
		EmailVerification: EmailVerificationConfig{
			APIURL:      apiURL,
			APIEndpoint: apiEndpoint,
			CLIPath:     cliPath,
			UseCLI:      useCLI,
		},
		Logging: LoggingConfig{
			Level:  logLevel,
			Format: logFormat,
		},
		RateLimit:               rateLimit,
		VerificationTimeout:     time.Duration(timeoutSeconds) * time.Second,
		MaxEmailPatterns:        maxPatterns,
		VerificationConcurrency: verificationConcurrency,
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetLogger() (*zap.Logger, error) {
	var config zap.Config

	if c.Logging.Format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	switch c.Logging.Level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return config.Build()
}
