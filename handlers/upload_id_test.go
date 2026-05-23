package handlers

import (
	"regexp"
	"testing"
)

func TestGenerateImageIDIsUniqueAndTimestampPrefixed(t *testing.T) {
	const count = 10000
	seen := make(map[string]struct{}, count)
	pattern := regexp.MustCompile(`^\d{8}_\d{6}_[a-f0-9]{32}$`)

	for i := 0; i < count; i++ {
		id, err := generateImageID()
		if err != nil {
			t.Fatalf("generateImageID returned error: %v", err)
		}
		if !pattern.MatchString(id) {
			t.Fatalf("generateImageID returned unexpected format: %q", id)
		}
		if _, exists := seen[id]; exists {
			t.Fatalf("generateImageID returned duplicate ID: %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestUploadResultIncludesImageID(t *testing.T) {
	result := UploadResult{ID: "20260524_153012_550e8400e29b41d4a716446655440000"}
	if result.ID == "" {
		t.Fatal("UploadResult should expose image ID")
	}
}
