package functions

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func CreateHash(message string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	hash := hex.EncodeToString(h.Sum(nil))
	return hash
}
