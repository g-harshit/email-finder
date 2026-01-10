package main

import (
	"email-finder/config"
	"email-finder/internal/handler"
	"email-finder/internal/resolver"
	"email-finder/internal/service"
	"email-finder/internal/verifier"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Initialize logger
	logger, err := cfg.GetLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("starting email finder service",
		zap.String("host", cfg.Server.Host),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize email verifier
	var emailVerifier verifier.Verifier
	if cfg.EmailVerification.UseCLI {
		logger.Info("using CLI verifier",
			zap.String("path", cfg.EmailVerification.CLIPath),
			zap.Int("concurrency", cfg.VerificationConcurrency),
		)
		emailVerifier = verifier.NewCLIVerifier(
			cfg.EmailVerification.CLIPath,
			cfg.VerificationTimeout,
			cfg.VerificationConcurrency,
			logger,
		)
	} else {
		logger.Info("using HTTP verifier",
			zap.String("url", cfg.EmailVerification.APIURL),
			zap.String("endpoint", cfg.EmailVerification.APIEndpoint),
			zap.Int("concurrency", cfg.VerificationConcurrency),
		)
		emailVerifier = verifier.NewHTTPVerifier(
			cfg.EmailVerification.APIURL,
			cfg.EmailVerification.APIEndpoint,
			cfg.VerificationTimeout,
			cfg.VerificationConcurrency,
			logger,
		)
	}

	// Initialize domain resolver
	domainResolver := resolver.NewDomainResolver(
		logger,
		cfg.VerificationTimeout,
	)

	// Initialize service
	emailFinderService := service.NewEmailFinderService(
		emailVerifier,
		domainResolver,
		logger,
		cfg.MaxEmailPatterns,
	)

	// Initialize handler
	emailHandler := handler.NewEmailHandler(emailFinderService, logger)

	// Setup router
	router := setupRouter(emailHandler, logger, cfg)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logger.Info("server starting", zap.String("address", addr))

	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal("server failed to start", zap.Error(err))
	}
}

func setupRouter(emailHandler *handler.EmailHandler, logger *zap.Logger, cfg *config.Config) *gin.Engine {
	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(ginLogger(logger))
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check
	router.GET("/health", emailHandler.HealthCheck)

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.POST("/find-email", emailHandler.FindEmail)
	}

	return router
}

func ginLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		logger.Info("HTTP request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
