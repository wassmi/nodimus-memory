package models

import (
	"crypto/sha256"
	"encoding/hex"
)

// ModelChecksums holds the checksums for the models.
var ModelChecksums = map[string]string{
	"bert-base-uncased": "0a6aa92c532b63c6b2256e4e6a703907b6543785d54588ac2f3237d84952d169",
}

// VerifyChecksum verifies the checksum of a model.
func VerifyChecksum(modelName string, data []byte) bool {
	hash := sha256.Sum256(data)
	return ModelChecksums[modelName] == hex.EncodeToString(hash[:])
}
