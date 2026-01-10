package handler

import (
	"email-finder/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EmailHandler handles HTTP requests for email finding
type EmailHandler struct {
	service *service.EmailFinderService
	logger  *zap.Logger
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(svc *service.EmailFinderService, logger *zap.Logger) *EmailHandler {
	return &EmailHandler{
		service: svc,
		logger:  logger,
	}
}

// FindEmail handles POST /api/v1/find-email
func (h *EmailHandler) FindEmail(c *gin.Context) {
	var req service.FindEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request. Please provide first_name, last_name, and company.",
			"details": err.Error(),
		})
		return
	}

	// Validate inputs
	if req.FirstName == "" || req.LastName == "" || req.Company == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "first_name, last_name, and company are required fields",
		})
		return
	}

	// Find emails
	result, err := h.service.FindEmails(req)
	if err != nil {
		h.logger.Error("failed to find emails", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process email search",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HealthCheck handles GET /health
func (h *EmailHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "email-finder",
	})
}
