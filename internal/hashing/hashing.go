package hashing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Calculate(data []byte, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data) // Write never returns error for hash.Hash

	sign := h.Sum(nil)
	return hex.EncodeToString(sign)
}
