package models

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyChecksum(t *testing.T) {
	// Test with a known good checksum
	modelName := "bert-base-uncased"
	data := []byte("This is a test string for bert-base-uncased.")
	hash := sha256.Sum256(data)
	correctChecksum := hex.EncodeToString(hash[:])
	ModelChecksums[modelName] = correctChecksum

	if !VerifyChecksum(modelName, data) {
		t.Errorf("VerifyChecksum failed: expected true for a matching checksum, got false")
	}

	// Test with a mismatched checksum
	mismatchedData := []byte("This is a different test string.")
	if VerifyChecksum(modelName, mismatchedData) {
		t.Errorf("VerifyChecksum failed: expected false for a mismatched checksum, got true")
	}

	// Test with an unknown model
	if VerifyChecksum("unknown-model", data) {
		t.Errorf("VerifyChecksum failed: expected false for an unknown model, got true")
	}
}