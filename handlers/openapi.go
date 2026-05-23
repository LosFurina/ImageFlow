package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/LosFurina/ImageFlow/auth"
	"github.com/LosFurina/ImageFlow/config"
	"github.com/LosFurina/ImageFlow/utils"
	"github.com/redis/go-redis/v9"
)

// RegisterOpenAPIRoutes registers all /openapi/* routes with AK/SK authentication.
func RegisterOpenAPIRoutes(rdb *redis.Client, cfg *config.Config) {
	http.HandleFunc("/openapi/upload", OpenAPIUploadHandler(rdb, cfg))
	http.HandleFunc("/openapi/images", OpenAPIListImagesHandler(rdb, cfg))
	http.HandleFunc("/openapi/delete", OpenAPIDeleteImageHandler(rdb, cfg))
	http.HandleFunc("/openapi/tags", OpenAPITagsHandler(rdb, cfg))
	http.HandleFunc("/openapi/config", OpenAPIConfigHandler(rdb, cfg))
	http.HandleFunc("/openapi/random", OpenAPIRandomHandler(rdb, cfg))
	http.HandleFunc("/openapi/debug/tags", OpenAPIDebugTagsHandler(rdb, cfg))
	http.HandleFunc("/openapi/cleanup", OpenAPICleanupHandler(rdb))
}

// OpenAPIUploadHandler godoc
// @Summary Upload images
// @Description Upload images through OpenAPI. Requires api:upload permission.
// @Tags OpenAPI
// @Accept multipart/form-data
// @Produce json
// @Param X-Access-Key header string true "Access Key"
// @Param X-Signature header string true "HMAC-SHA256 signature"
// @Param X-Timestamp header string true "Unix timestamp"
// @Param images formData file true "Image files"
// @Param tags formData string false "Comma-separated tags"
// @Param expiryMinutes formData int false "Expiration in minutes"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /openapi/upload [post]
func OpenAPIUploadHandler(rdb *redis.Client, cfg *config.Config) http.HandlerFunc {
	return auth.AKSKAuthMiddleware(rdb, auth.PermissionUpload, UploadHandler(cfg))
}

// OpenAPIListImagesHandler godoc
// @Summary List images
// @Description List uploaded images. Requires api:images permission.
// @Tags OpenAPI
// @Produce json
// @Param X-Access-Key header string true "Access Key"
// @Param X-Signature header string true "HMAC-SHA256 signature"
// @Param X-Timestamp header string true "Unix timestamp"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Param tag query string false "Filter by tag"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /openapi/images [get]
func OpenAPIListImagesHandler(rdb *redis.Client, cfg *config.Config) http.HandlerFunc {
	return auth.AKSKAuthMiddleware(rdb, auth.PermissionImages, PublicListImagesHandler(cfg))
}

// OpenAPIDeleteImageHandler godoc
// @Summary Delete image
// @Description Delete an image. Requires api:delete permission.
// @Tags OpenAPI
// @Accept json
// @Produce json
// @Param X-Access-Key header string true "Access Key"
// @Param X-Signature header string true "HMAC-SHA256 signature"
// @Param X-Timestamp header string true "Unix timestamp"
// @Param request body map[string]string true "Delete request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /openapi/delete [post]
func OpenAPIDeleteImageHandler(rdb *redis.Client, cfg *config.Config) http.HandlerFunc {
	return auth.AKSKAuthMiddleware(rdb, auth.PermissionDelete, DeleteImageHandler(cfg))
}

// OpenAPITagsHandler godoc
// @Summary List tags
// @Description List available image tags. Requires api:tags permission.
// @Tags OpenAPI
// @Produce json
// @Param X-Access-Key header string true "Access Key"
// @Param X-Signature header string true "HMAC-SHA256 signature"
// @Param X-Timestamp header string true "Unix timestamp"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /openapi/tags [get]
func OpenAPITagsHandler(rdb *redis.Client, cfg *config.Config) http.HandlerFunc {
	return auth.AKSKAuthMiddleware(rdb, auth.PermissionTags, TagsHandler(cfg))
}

// OpenAPIConfigHandler godoc
// @Summary Get client configuration
// @Description Get public client configuration. Requires api:config permission.
// @Tags OpenAPI
// @Produce json
// @Param X-Access-Key header string true "Access Key"
// @Param X-Signature header string true "HMAC-SHA256 signature"
// @Param X-Timestamp header string true "Unix timestamp"
// @Success 200 {object} config.ClientConfig
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /openapi/config [get]
func OpenAPIConfigHandler(rdb *redis.Client, cfg *config.Config) http.HandlerFunc {
	return auth.AKSKAuthMiddleware(rdb, auth.PermissionConfig, ConfigHandler(cfg))
}

// OpenAPIRandomHandler godoc
// @Summary Get random image
// @Description Get a random image with optional filters. Requires api:random permission.
// @Tags OpenAPI
// @Produce json
// @Param X-Access-Key header string true "Access Key"
// @Param X-Signature header string true "HMAC-SHA256 signature"
// @Param X-Timestamp header string true "Unix timestamp"
// @Param tag query string false "Single tag filter"
// @Param tags query string false "Comma-separated tag filters"
// @Param exclude query string false "Comma-separated excluded tags"
// @Param orientation query string false "portrait or landscape"
// @Param format query string false "avif, webp, or original"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /openapi/random [get]
func OpenAPIRandomHandler(rdb *redis.Client, cfg *config.Config) http.HandlerFunc {
	if cfg.StorageType == config.StorageTypeS3 {
		return auth.AKSKAuthMiddleware(rdb, auth.PermissionRandom, RandomImageHandler(utils.S3Client, cfg))
	}
	return auth.AKSKAuthMiddleware(rdb, auth.PermissionRandom, LocalRandomImageHandler(cfg))
}

// OpenAPIDebugTagsHandler godoc
// @Summary Debug tags
// @Description Debug tag indexes. Requires api:debug permission.
// @Tags OpenAPI
// @Produce json
// @Param X-Access-Key header string true "Access Key"
// @Param X-Signature header string true "HMAC-SHA256 signature"
// @Param X-Timestamp header string true "Unix timestamp"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /openapi/debug/tags [get]
func OpenAPIDebugTagsHandler(rdb *redis.Client, cfg *config.Config) http.HandlerFunc {
	return auth.AKSKAuthMiddleware(rdb, auth.PermissionDebug, DebugTagsHandler(cfg))
}

// OpenAPICleanupHandler godoc
// @Summary Trigger cleanup
// @Description Trigger expired image cleanup. Requires api:cleanup permission.
// @Tags OpenAPI
// @Produce json
// @Param X-Access-Key header string true "Access Key"
// @Param X-Signature header string true "HMAC-SHA256 signature"
// @Param X-Timestamp header string true "Unix timestamp"
// @Success 200 {object} map[string]string
// @Failure 401 {object} errors.ErrorResponse
// @Failure 403 {object} errors.ErrorResponse
// @Router /openapi/cleanup [post]
func OpenAPICleanupHandler(rdb *redis.Client) http.HandlerFunc {
	return auth.AKSKAuthMiddleware(rdb, auth.PermissionCleanup, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		utils.TriggerCleanup()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Cleanup process triggered",
		})
	})
}
