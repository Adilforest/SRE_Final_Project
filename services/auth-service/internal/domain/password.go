package domain

import (
	"golang.org/x/crypto/bcrypt"
	"github.com/sirupsen/logrus"
	"time"
)

const bcryptCost = bcrypt.DefaultCost

func HashPassword(password string) (string, error) {
    start := time.Now()
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    if err != nil {
        logrus.WithError(err).Error("Failed to hash password")
        return "", err
    }
    
    logrus.WithFields(logrus.Fields{
        "duration": time.Since(start),
    }).Debug("Password hashed successfully")
    
    return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        logrus.WithError(err).Warn("Password validation failed")
        return false
    }
    return true
}