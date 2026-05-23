package auth

import (
	"strconv"
	"testing"
	"time"
)

func TestEffectivePermissions(t *testing.T) {
	perms := EffectivePermissions("reader", []string{PermissionUpload})

	if !HasPermission(perms, PermissionRandom) {
		t.Fatalf("reader role should include %s", PermissionRandom)
	}
	if !HasPermission(perms, PermissionUpload) {
		t.Fatalf("custom permissions should include %s", PermissionUpload)
	}
	if HasPermission(perms, PermissionDelete) {
		t.Fatalf("reader + upload should not include %s", PermissionDelete)
	}
}

func TestSignatureVerification(t *testing.T) {
	sk := "secret-key"
	method := "POST"
	path := "/openapi/upload"
	timestamp := time.Now().Format("20060102150405")
	body := []byte(`{"hello":"world"}`)

	stringToSign := BuildStringToSign(method, path, timestamp, body)
	sig := ComputeSignature(sk, stringToSign)

	if !VerifySignature(sk, method, path, timestamp, body, sig) {
		t.Fatal("expected signature verification to pass")
	}

	if VerifySignature(sk, method, path, timestamp, []byte(`{"tampered":true}`), sig) {
		t.Fatal("expected tampered body signature verification to fail")
	}
}

func TestTimestampValidation(t *testing.T) {
	valid := time.Now().Unix()
	if err := ValidateTimestamp(formatUnix(valid)); err != nil {
		t.Fatalf("expected current timestamp to be valid: %v", err)
	}

	expired := time.Now().Add(-10 * time.Minute).Unix()
	if err := ValidateTimestamp(formatUnix(expired)); err == nil {
		t.Fatal("expected expired timestamp to fail")
	}
}

func TestSKEncryptionRoundTrip(t *testing.T) {
	InitSKEncryption("test-admin-api-key")
	original := "very-secret-key"

	encrypted, err := encryptSK(original)
	if err != nil {
		t.Fatalf("encryptSK failed: %v", err)
	}
	if encrypted == original {
		t.Fatal("encrypted SK should not equal plaintext")
	}

	decrypted, err := decryptSK(encrypted)
	if err != nil {
		t.Fatalf("decryptSK failed: %v", err)
	}
	if decrypted != original {
		t.Fatalf("expected decrypted SK %q, got %q", original, decrypted)
	}
}

func formatUnix(ts int64) string {
	return strconv.FormatInt(ts, 10)
}
