package handlers

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/LosFurina/ImageFlow/config"
	"github.com/LosFurina/ImageFlow/utils/errors"
	"github.com/LosFurina/ImageFlow/utils/logger"
	"go.uber.org/zap"
)

// maskAPIKey returns a masked version of the API key for safe logging
// Shows first 4 characters followed by asterisks
func maskAPIKey(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return key[:4] + "****"
}

// AuthResponse represents the response for API key validation
type AuthResponse struct {
	Valid bool   `json:"valid"`           // Whether the API key is valid
	Error string `json:"error,omitempty"` // Error message if validation fails
}

// ValidateAPIKey provides an endpoint to validate API keys
func ValidateAPIKey(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Get API key from request header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errors.WriteError(w, errors.ErrInvalidAPIKey)
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errors.WriteError(w, errors.ErrInvalidAPIKey)
			return
		}

		providedKey := parts[1]

		// Validate API key using constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(cfg.APIKey)) == 1 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"valid":true}`))
			logger.Debug("API key validated successfully")
		} else {
			errors.WriteError(w, errors.ErrInvalidAPIKey)
			logger.Warn("API key validation failed",
				zap.String("masked_key", maskAPIKey(providedKey)))
		}
	}
}

// RequireAPIKey middleware to validate API key before processing requests
func RequireAPIKey(cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get API key from request header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errors.WriteError(w, errors.ErrInvalidAPIKey)
			logger.Warn("Missing API key",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errors.WriteError(w, errors.ErrInvalidAPIKey)
			logger.Warn("Invalid Authorization header format",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			return
		}

		// Validate API key using constant-time comparison to prevent timing attacks
		providedKey := parts[1]
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(cfg.APIKey)) != 1 {
			errors.WriteError(w, errors.ErrInvalidAPIKey)
			logger.Warn("API key validation failed",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("masked_key", maskAPIKey(providedKey)))
			return
		}

		// If API key is valid, proceed to next handler
		next(w, r)
	}
}
