package service

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"gophkeeper/pkg/crypt"
)

type Auth struct {
	UserID         int64
	PrivateKey     *rsa.PrivateKey
	PublicKeyBytes []byte
	Hash           []byte
	Fingerprint    string
}

func NewAuth() *Auth {
	return &Auth{}
}

func (s *Auth) SetPrivateKeyFromPath(privateKeyPath string) error {
	privateKey, err := crypt.ParsePrivateKey(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	return s.SetPrivateKey(privateKey)
}

func (s *Auth) SetPrivateKey(privateKey *rsa.PrivateKey) error {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}
	s.PrivateKey = privateKey
	s.PublicKeyBytes = publicKeyBytes
	s.Hash = crypt.PublicKeyHash(s.PublicKeyBytes)
	s.Fingerprint = crypt.FingerprintFromHash(s.Hash)
	return nil
}

func (s *Auth) SetUserID(userID int64) {
	s.UserID = userID
}
