package main

import (
	"crypto/sha512"
	"encoding/hex"

	"github.com/google/uuid"
)

func Hash(str string) string {
	hash := sha512.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum(nil))
}

func GenerateUUID() string {
	return uuid.New().String()
}
