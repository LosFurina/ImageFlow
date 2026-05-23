package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/LosFurina/ImageFlow/auth"
	"github.com/LosFurina/ImageFlow/config"
	"github.com/LosFurina/ImageFlow/utils/errors"
	"github.com/LosFurina/ImageFlow/utils/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AKSKCreateRequest is the request body for creating a new AK/SK pair
type AKSKCreateRequest struct {
	Name              string   `json:"name"`               // Required: user-friendly name
	Description       string   `json:"description"`        // Optional: description
	Role              string   `json:"role"`               // Required: reader, writer, admin
	CustomPermissions []string `json:"custom_permissions"` // Optional: additional permissions
}

// AKSKUpdateRequest is the request body for updating an AK/SK entry
type AKSKUpdateRequest struct {
	AccessKey         string   `json:"access_key"`         // Required: the AK to update
	Name              *string  `json:"name"`               // Optional: new name
	Description       *string  `json:"description"`        // Optional: new description
	Role              *string  `json:"role"`               // Optional: new role
	CustomPermissions []string `json:"custom_permissions"` // Optional: replace custom permissions (null = no change)
	Enabled           *bool    `json:"enabled"`            // Optional: enable/disable
}

// AKSKDeleteRequest is the request body for deleting an AK/SK entry
type AKSKDeleteRequest struct {
	AccessKey string `json:"access_key"` // Required: the AK to delete
}

// AKSKRotateRequest is the request body for rotating an SK
type AKSKRotateRequest struct {
	AccessKey string `json:"access_key"` // Required: the AK whose SK to rotate
}

// RegisterAKSKAdminRoutes registers admin API routes for AK/SK management
func RegisterAKSKAdminRoutes(rdb *redis.Client, cfg *config.Config) {
	// List all AK/SK entries
	http.HandleFunc("/api/admin/aksk/list", RequireAPIKey(cfg, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			errors.HandleError(w, errors.ErrInvalidParam, "Method not allowed", nil)
			return
		}

		entries, err := auth.ListAKSK(rdb)
		if err != nil {
			errors.HandleError(w, errors.ErrInternal, "Failed to list AK/SK entries", err.Error())
			logger.Error("Failed to list AK/SK entries", zap.Error(err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"entries": entries,
			"count":   len(entries),
		})
	}))

	// Create a new AK/SK pair
	http.HandleFunc("/api/admin/aksk/create", RequireAPIKey(cfg, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errors.HandleError(w, errors.ErrInvalidParam, "Method not allowed", nil)
			return
		}

		var req AKSKCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errors.HandleError(w, errors.ErrInvalidParam, "Invalid request body", err.Error())
			return
		}

		// Validate required fields
		if req.Name == "" {
			errors.HandleError(w, errors.ErrInvalidParam, "Name is required", nil)
			return
		}
		if req.Role == "" || !auth.IsValidRole(req.Role) {
			errors.HandleError(w, errors.ErrInvalidParam, "Valid role is required (reader, writer, admin)", nil)
			return
		}

		// Validate custom permissions
		for _, p := range req.CustomPermissions {
			if !auth.IsValidPermission(p) {
				errors.HandleError(w, errors.ErrInvalidParam, "Invalid permission: "+p, nil)
				return
			}
		}

		result, err := auth.CreateAKSK(rdb, req.Name, req.Description, req.Role, req.CustomPermissions)
		if err != nil {
			errors.HandleError(w, errors.ErrInternal, "Failed to create AK/SK", err.Error())
			logger.Error("Failed to create AK/SK", zap.Error(err))
			return
		}

		logger.Info("AK/SK created",
			zap.String("ak", result.AccessKey[:4]+"****"),
			zap.String("role", req.Role))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)
	}))

	// Update an AK/SK entry
	http.HandleFunc("/api/admin/aksk/update", RequireAPIKey(cfg, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			errors.HandleError(w, errors.ErrInvalidParam, "Method not allowed", nil)
			return
		}

		var req AKSKUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errors.HandleError(w, errors.ErrInvalidParam, "Invalid request body", err.Error())
			return
		}

		if req.AccessKey == "" {
			errors.HandleError(w, errors.ErrInvalidParam, "Access key is required", nil)
			return
		}

		// Load existing entry
		entry, err := auth.LoadAKSK(rdb, req.AccessKey)
		if err != nil {
			errors.HandleError(w, errors.ErrInternal, "Failed to load AK/SK entry", err.Error())
			return
		}
		if entry == nil {
			errors.HandleError(w, errors.ErrNotFound, "AK/SK entry not found", nil)
			return
		}

		// Apply updates
		if req.Name != nil {
			entry.Name = *req.Name
		}
		if req.Description != nil {
			entry.Description = *req.Description
		}
		if req.Role != nil {
			if !auth.IsValidRole(*req.Role) {
				errors.HandleError(w, errors.ErrInvalidParam, "Invalid role", nil)
				return
			}
			entry.Role = *req.Role
		}
		if req.CustomPermissions != nil {
			for _, p := range req.CustomPermissions {
				if !auth.IsValidPermission(p) {
					errors.HandleError(w, errors.ErrInvalidParam, "Invalid permission: "+p, nil)
					return
				}
			}
			entry.CustomPermissions = req.CustomPermissions
		}
		if req.Enabled != nil {
			entry.Enabled = *req.Enabled
		}

		if err := auth.SaveAKSK(rdb, req.AccessKey, entry); err != nil {
			errors.HandleError(w, errors.ErrInternal, "Failed to update AK/SK entry", err.Error())
			logger.Error("Failed to update AK/SK entry", zap.Error(err))
			return
		}

		logger.Info("AK/SK updated",
			zap.String("ak", req.AccessKey[:4]+"****"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "AK/SK entry updated",
		})
	}))

	// Delete an AK/SK entry
	http.HandleFunc("/api/admin/aksk/delete", RequireAPIKey(cfg, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			errors.HandleError(w, errors.ErrInvalidParam, "Method not allowed", nil)
			return
		}

		var req AKSKDeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errors.HandleError(w, errors.ErrInvalidParam, "Invalid request body", err.Error())
			return
		}

		if req.AccessKey == "" {
			errors.HandleError(w, errors.ErrInvalidParam, "Access key is required", nil)
			return
		}

		// Check if entry exists
		entry, err := auth.LoadAKSK(rdb, req.AccessKey)
		if err != nil {
			errors.HandleError(w, errors.ErrInternal, "Failed to load AK/SK entry", err.Error())
			return
		}
		if entry == nil {
			errors.HandleError(w, errors.ErrNotFound, "AK/SK entry not found", nil)
			return
		}

		if err := auth.DeleteAKSK(rdb, req.AccessKey); err != nil {
			errors.HandleError(w, errors.ErrInternal, "Failed to delete AK/SK entry", err.Error())
			logger.Error("Failed to delete AK/SK entry", zap.Error(err))
			return
		}

		logger.Info("AK/SK deleted",
			zap.String("ak", req.AccessKey[:4]+"****"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "AK/SK entry deleted",
		})
	}))

	// Rotate SK for an AK
	http.HandleFunc("/api/admin/aksk/rotate", RequireAPIKey(cfg, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			errors.HandleError(w, errors.ErrInvalidParam, "Method not allowed", nil)
			return
		}

		var req AKSKRotateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errors.HandleError(w, errors.ErrInvalidParam, "Invalid request body", err.Error())
			return
		}

		if req.AccessKey == "" {
			errors.HandleError(w, errors.ErrInvalidParam, "Access key is required", nil)
			return
		}

		result, err := auth.RotateSK(rdb, req.AccessKey)
		if err != nil {
			errors.HandleError(w, errors.ErrInternal, "Failed to rotate SK", err.Error())
			logger.Error("Failed to rotate SK", zap.Error(err))
			return
		}

		logger.Info("SK rotated",
			zap.String("ak", result.AccessKey[:4]+"****"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}))
}
