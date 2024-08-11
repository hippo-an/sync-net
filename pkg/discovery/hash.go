package discovery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
)

const (
	key = "d32d8715eb065833a7cbd2ae6f6e845693780c17ebaae9bc889b4d97bb2c30f8"
)

var (
	ErrInvalidKey = errors.New("invalid key error")
	h             = generateHash()
)

func validateHash(hash string) error {
	if h != hash {
		log.Println("Hash mismatch! The message may have been tampered with.")
		return ErrInvalidKey
	}

	return nil
}

func generateHash() string {
	hs := hmac.New(sha256.New, []byte(key))
	return hex.EncodeToString(hs.Sum([]byte(key)))
}
