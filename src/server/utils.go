package main

import (
	"crypto/sha512"
	"encoding/hex"

	"github.com/google/uuid"
)

func HashPassword(password string) string {
	hash := sha512.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

func GenerateUUID() string {
	return uuid.New().String()
}
