package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LosFurina/ImageFlow/utils/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// Redis key prefix for AK/SK entries
	akskKeyPrefix = "imageflow:aksk:"

	// AK length (base62 characters)
	akLength = 20

	// SK length (base62 characters)
	skLength = 40
)

// base62 characters for generating AK/SK
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// AKSKEntry represents an AK/SK credential entry stored in Redis
type AKSKEntry struct {
	SKEncrypted       string   `json:"sk_encrypted"`       // AES-GCM encrypted SK (decrypted at runtime for verification)
	Name              string   `json:"name"`               // User-friendly name
	Description       string   `json:"description"`        // Optional description
	Role              string   `json:"role"`               // Preset role: reader, writer, admin
	CustomPermissions []string `json:"custom_permissions"` // Additional permissions beyond role
	CreatedAt         int64    `json:"created_at"`         // Unix timestamp of creation
	Enabled           bool     `json:"enabled"`            // Whether this AK/SK is active
}

// AKSKCreateResult is returned when creating a new AK/SK pair
type AKSKCreateResult struct {
	AccessKey string `json:"access_key"` // The generated AK (plaintext)
	SecretKey string `json:"secret_key"` // The generated SK (plaintext, shown only once)
}

// generateRandomString generates a cryptographically secure random base62 string
func generateRandomString(length int) (string, error) {
	result := make([]byte, length)
	randomBytes := make([]byte, length)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	for i := 0; i < length; i++ {
		result[i] = base62Chars[randomBytes[i]%byte(len(base62Chars))]
	}

	return string(result), nil
}

// GenerateAK generates a new Access Key
func GenerateAK() (string, error) {
	return generateRandomString(akLength)
}

// GenerateSK generates a new Secret Key
func GenerateSK() (string, error) {
	return generateRandomString(skLength)
}

// SaveAKSK stores an AK/SK entry in Redis
func SaveAKSK(rdb *redis.Client, ak string, entry *AKSKEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal AKSK entry: %w", err)
	}

	key := akskKeyPrefix + ak
	if err := rdb.Set(context.Background(), key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save AKSK entry to Redis: %w", err)
	}

	logger.Info("AK/SK entry saved",
		zap.String("ak", ak[:4]+"****"),
		zap.String("role", entry.Role))

	return nil
}

// LoadAKSK loads an AK/SK entry from Redis
func LoadAKSK(rdb *redis.Client, ak string) (*AKSKEntry, error) {
	key := akskKeyPrefix + ak
	data, err := rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to load AKSK entry from Redis: %w", err)
	}

	var entry AKSKEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal AKSK entry: %w", err)
	}

	return &entry, nil
}

// DeleteAKSK removes an AK/SK entry from Redis
func DeleteAKSK(rdb *redis.Client, ak string) error {
	key := akskKeyPrefix + ak
	if err := rdb.Del(context.Background(), key).Err(); err != nil {
		return fmt.Errorf("failed to delete AKSK entry from Redis: %w", err)
	}

	logger.Info("AK/SK entry deleted",
		zap.String("ak", ak[:4]+"****"))

	return nil
}

// ListAKSK lists all AK/SK entries from Redis
func ListAKSK(rdb *redis.Client) ([]AKSKListEntry, error) {
	var cursor uint64
	var entries []AKSKListEntry
	pattern := akskKeyPrefix + "*"

	for {
		keys, nextCursor, err := rdb.Scan(context.Background(), cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan AKSK keys: %w", err)
		}

		for _, key := range keys {
			// Extract AK from the full key
			ak := key[len(akskKeyPrefix):]

			data, err := rdb.Get(context.Background(), key).Bytes()
			if err != nil {
				logger.Warn("Failed to load AKSK entry during list",
					zap.String("ak", ak[:4]+"****"),
					zap.Error(err))
				continue
			}

			var entry AKSKEntry
			if err := json.Unmarshal(data, &entry); err != nil {
				logger.Warn("Failed to unmarshal AKSK entry during list",
					zap.String("ak", ak[:4]+"****"),
					zap.Error(err))
				continue
			}

			entries = append(entries, AKSKListEntry{
				AccessKey:         ak,
				Name:              entry.Name,
				Description:       entry.Description,
				Role:              entry.Role,
				CustomPermissions: entry.CustomPermissions,
				CreatedAt:         entry.CreatedAt,
				Enabled:           entry.Enabled,
			})
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return entries, nil
}

// AKSKListEntry is the public representation of an AK/SK entry (no SK hash)
type AKSKListEntry struct {
	AccessKey         string   `json:"access_key"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Role              string   `json:"role"`
	CustomPermissions []string `json:"custom_permissions"`
	CreatedAt         int64    `json:"created_at"`
	Enabled           bool     `json:"enabled"`
}

// CreateAKSK generates a new AK/SK pair and stores it in Redis
func CreateAKSK(rdb *redis.Client, name, description, role string, customPermissions []string) (*AKSKCreateResult, error) {
	ak, err := GenerateAK()
	if err != nil {
		return nil, fmt.Errorf("failed to generate AK: %w", err)
	}

	sk, err := GenerateSK()
	if err != nil {
		return nil, fmt.Errorf("failed to generate SK: %w", err)
	}

	encryptedSK, err := encryptSK(sk)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt SK: %w", err)
	}

	entry := &AKSKEntry{
		SKEncrypted:       encryptedSK,
		Name:              name,
		Description:       description,
		Role:              role,
		CustomPermissions: customPermissions,
		CreatedAt:         time.Now().Unix(),
		Enabled:           true,
	}

	if err := SaveAKSK(rdb, ak, entry); err != nil {
		return nil, fmt.Errorf("failed to save new AKSK entry: %w", err)
	}

	return &AKSKCreateResult{
		AccessKey: ak,
		SecretKey: sk,
	}, nil
}

// RotateSK generates a new SK for an existing AK and updates the Redis entry
func RotateSK(rdb *redis.Client, ak string) (*AKSKCreateResult, error) {
	entry, err := LoadAKSK(rdb, ak)
	if err != nil {
		return nil, fmt.Errorf("failed to load AKSK entry for rotation: %w", err)
	}
	if entry == nil {
		return nil, fmt.Errorf("AK/SK entry not found: %s", ak[:4]+"****")
	}

	newSK, err := GenerateSK()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new SK: %w", err)
	}

	encryptedSK, err := encryptSK(newSK)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt new SK: %w", err)
	}

	entry.SKEncrypted = encryptedSK

	if err := SaveAKSK(rdb, ak, entry); err != nil {
		return nil, fmt.Errorf("failed to update AKSK entry after rotation: %w", err)
	}

	logger.Info("SK rotated",
		zap.String("ak", ak[:4]+"****"))

	return &AKSKCreateResult{
		AccessKey: ak,
		SecretKey: newSK,
	}, nil
}
