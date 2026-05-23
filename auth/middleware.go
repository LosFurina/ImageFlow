package auth

import (
	"bytes"
	"io"
	"net/http"

	"github.com/LosFurina/ImageFlow/utils/errors"
	"github.com/LosFurina/ImageFlow/utils/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AKSKAuthMiddleware validates AK/SK HMAC signature and checks permissions
func AKSKAuthMiddleware(rdb *redis.Client, requiredPermission string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract AK/SK headers
		accessKey := r.Header.Get("X-Access-Key")
		signature := r.Header.Get("X-Signature")
		timestamp := r.Header.Get("X-Timestamp")

		if accessKey == "" || signature == "" || timestamp == "" {
			errors.WriteError(w, errors.ErrAKSKInvalidCreds)
			logger.Warn("Missing AK/SK headers",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			return
		}

		// Validate timestamp
		if err := ValidateTimestamp(timestamp); err != nil {
			errors.WriteError(w, errors.NewError(errors.ErrUnauthorized, "Request timestamp expired", err.Error()))
			logger.Warn("Timestamp validation failed",
				zap.String("path", r.URL.Path),
				zap.String("timestamp", timestamp),
				zap.Error(err))
			return
		}

		// Load AK/SK entry from Redis
		entry, err := LoadAKSK(rdb, accessKey)
		if err != nil {
			errors.WriteError(w, errors.ErrServerError)
			logger.Error("Failed to load AKSK entry",
				zap.String("ak", maskAK(accessKey)),
				zap.Error(err))
			return
		}

		if entry == nil {
			errors.WriteError(w, errors.ErrAKSKInvalidCreds)
			logger.Warn("AK not found",
				zap.String("ak", maskAK(accessKey)),
				zap.String("path", r.URL.Path))
			return
		}

		// Check if AK is enabled
		if !entry.Enabled {
			errors.WriteError(w, errors.ErrAKSKAccountDisabled)
			logger.Warn("AK is disabled",
				zap.String("ak", maskAK(accessKey)),
				zap.String("path", r.URL.Path))
			return
		}

		// Decrypt SK for signature verification
		sk, err := decryptSK(entry.SKEncrypted)
		if err != nil {
			errors.WriteError(w, errors.ErrServerError)
			logger.Error("Failed to decrypt SK for verification",
				zap.String("ak", maskAK(accessKey)),
				zap.Error(err))
			return
		}

		// Read request body for signature verification
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			errors.WriteError(w, errors.ErrServerError)
			logger.Error("Failed to read request body for signature verification",
				zap.Error(err))
			return
		}
		// Restore body so downstream handlers can read it
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// Verify HMAC signature
		if !VerifySignature(sk, r.Method, r.URL.Path, timestamp, bodyBytes, signature) {
			errors.WriteError(w, errors.ErrAKSKInvalidCreds)
			logger.Warn("Signature verification failed",
				zap.String("ak", maskAK(accessKey)),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			return
		}

		// Check permissions
		effectivePerms := EffectivePermissions(entry.Role, entry.CustomPermissions)
		if !HasPermission(effectivePerms, requiredPermission) {
			errors.WriteError(w, errors.NewError(errors.ErrForbidden, "Insufficient permissions",
				map[string]interface{}{
					"required":  requiredPermission,
					"role":      entry.Role,
					"effective": effectivePerms,
				}))
			logger.Warn("Permission denied",
				zap.String("ak", maskAK(accessKey)),
				zap.String("required", requiredPermission),
				zap.Strings("effective", effectivePerms))
			return
		}

		// All checks passed, proceed to handler
		next(w, r)
	}
}

// maskAK returns a masked version of the AK for safe logging
func maskAK(ak string) string {
	if len(ak) <= 4 {
		return "****"
	}
	return ak[:4] + "****"
}
