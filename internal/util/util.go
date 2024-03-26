package util

import (
	"crypto/rand"
	"math/big"

	"go.uber.org/zap"

	"github.com/pluhe7/gophermart/internal/logger"
)

func GenerateRandomString(length int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	b := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			logger.Log.Error("generate random string: rand int error", zap.Error(err))
			return ""
		}

		b[i] = letters[num.Int64()]
	}

	return string(b)
}
