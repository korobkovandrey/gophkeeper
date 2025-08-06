package service

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"gophkeeper/pkg/crypt"
)

type Key struct {
	UserID         int64
	PrivateKey     *rsa.PrivateKey
	PublicKeyBytes []byte
	Hash           []byte
	Fingerprint    string
}

func NewKey() *Key {
	return &Key{}
}

func (s *Key) SetPrivateKeyFromPath(privateKeyPath string) error {
	privateKey, err := crypt.ParsePrivateKey(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	return s.SetPrivateKey(privateKey)
}

func (s *Key) SetPrivateKey(privateKey *rsa.PrivateKey) error {
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

func (s *Key) SetUserID(userID int64) {
	s.UserID = userID
}
