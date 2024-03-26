package app

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt generage hash: %w", err)
	}

	return string(hash), nil
}

func checkPassword(hash, raw string) bool {
	if len(raw) == 0 || len(hash) == 0 {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw))
	return err == nil
}
