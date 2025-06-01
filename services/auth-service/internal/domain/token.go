package domain

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/sirupsen/logrus"
 
)

var ErrTokenGeneration = errors.New("failed to generate token")

func GenerateToken() (string, error) {
    b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        logrus.WithError(err).Error("Failed to generate token")
        return "", ErrTokenGeneration
    }

    token := base64.URLEncoding.EncodeToString(b)
    logrus.WithField("token", token[:8]+"...").Debug("Token generated")

    return token, nil
}

