package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// GenerateNewPrivateKey generates a new Private Key
func GenerateNewPrivateKey() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	return privateKey, err
}

// PrivateKeyToString converts a RSA private Key to a string
func PrivateKeyToString(privateKey *rsa.PrivateKey) string {
	keyBytes := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	return string(keyBytes[:])
}

// StringToPrivateKey converts a RSA Private Key String to a private KEy
func StringToPrivateKey(keyString string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(keyString))

	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("The key format is invalid.")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
