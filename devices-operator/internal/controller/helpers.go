package controller

import (
	"crypto/rand"
	"math/big"
)

func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		// ! c macro interpolation does not support some characters in !@#$%^&*()-_=+[]{}|;:,.<>?/~"
	password := make([]byte, length)
	for i := range password {
		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[randomIndex.Int64()]
	}
	return string(password)
}
