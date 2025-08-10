package crypt

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

func PublicKeyHash(publicKeyBytes []byte) []byte {
	hash := sha256.Sum256(publicKeyBytes)
	return hash[:]
}

func FingerprintFromHash(hash []byte) string {
	return base64.StdEncoding.EncodeToString(hash)
}

func Fingerprint(publicKeyBytes []byte) string {
	return FingerprintFromHash(PublicKeyHash(publicKeyBytes))
}

func RSAPublicKeyFromBytes(publicKeyBytes []byte) (*rsa.PublicKey, error) {
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key format")
	}
	return rsaPublicKey, nil
}

func EncodePublicKey(publicKeyBytes []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))
}

func DecodePublicKey(publicKeyStr string) ([]byte, error) {
	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, fmt.Errorf("invalid public key format")
	}
	return block.Bytes, nil
}

// ParsePrivateKey parses a private key from a file
func ParsePrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	key, err := ssh.ParseRawPrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("raw: %w", err)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not an RSA key")
	}
	if err = rsaKey.Validate(); err != nil {
		return nil, fmt.Errorf("private key is not valid: %w", err)
	}
	return rsaKey, nil
}
