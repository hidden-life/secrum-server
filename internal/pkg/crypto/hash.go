package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha256Hex(str string) string {
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:])
}
