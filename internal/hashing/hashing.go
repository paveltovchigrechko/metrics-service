package hashing

import (
	"crypto/hmac"
	"crypto/sha256"
)

func Calculate(data []byte, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data) // Write never returns error for hash.Hash

	sign := h.Sum(nil)
	return sign
}
