package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	// MaxTimestampSkew is the maximum allowed time difference between request timestamp and server time
	MaxTimestampSkew = 5 * time.Minute
)

// BuildStringToSign constructs the string to be signed
// Format: Method + "\n" + Path + "\n" + Timestamp + "\n" + SHA256(Body)
func BuildStringToSign(method, path, timestamp string, body []byte) string {
	bodyHash := SHA256Body(body)
	return fmt.Sprintf("%s\n%s\n%s\n%s", method, path, timestamp, bodyHash)
}

// SHA256Body computes the SHA256 hash of the request body
func SHA256Body(body []byte) string {
	if len(body) == 0 {
		body = []byte("")
	}
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// ComputeSignature computes the HMAC-SHA256 signature
func ComputeSignature(sk, stringToSign string) string {
	mac := hmac.New(sha256.New, []byte(sk))
	mac.Write([]byte(stringToSign))
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies the provided signature against the computed one
// Uses constant-time comparison to prevent timing attacks
func VerifySignature(sk, method, path, timestamp string, body []byte, providedSig string) bool {
	stringToSign := BuildStringToSign(method, path, timestamp, body)
	expectedSig := ComputeSignature(sk, stringToSign)
	return hmac.Equal([]byte(expectedSig), []byte(providedSig))
}

// ValidateTimestamp checks if the request timestamp is within the allowed skew
func ValidateTimestamp(timestampStr string) error {
	ts, err := parseTimestamp(timestampStr)
	if err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	now := time.Now().Unix()
	diff := now - ts
	if diff < 0 {
		diff = -diff
	}

	if time.Duration(diff)*time.Second > MaxTimestampSkew {
		return fmt.Errorf("request timestamp expired (skew: %v)", time.Duration(diff)*time.Second)
	}

	return nil
}

// parseTimestamp parses a Unix timestamp string
func parseTimestamp(s string) (int64, error) {
	var ts int64
	_, err := fmt.Sscanf(s, "%d", &ts)
	return ts, err
}
