package crypt

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	AESKeySize    = 32
	Int64BytesLen = 8
)

func HashData(data ...[]byte) []byte {
	hasher := sha256.New()
	for _, d := range data {
		hasher.Write(d)
	}
	return hasher.Sum(nil)
}

func VerifyPSSWithTimestamp(rsaPublicKey *rsa.PublicKey, signature []byte, timestamp int64, data ...[]byte) bool {
	timestampBytes := make([]byte, Int64BytesLen)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	return rsa.VerifyPSS(rsaPublicKey, crypto.SHA256, HashData(append(data, timestampBytes)...), signature, nil) == nil
}

func VerifyPSSWithTimestampAndUserID(rsaPublicKey *rsa.PublicKey, signature []byte, userID, timestamp int64, data ...[]byte) bool {
	userIDBytes := make([]byte, Int64BytesLen)
	binary.LittleEndian.PutUint64(userIDBytes, uint64(userID))
	return VerifyPSSWithTimestamp(rsaPublicKey, signature, timestamp, HashData(append(data, userIDBytes)...))
}

func SignPSSWithTimestamp(privateKey *rsa.PrivateKey, timestamp int64, data ...[]byte) ([]byte, error) {
	timestampBytes := make([]byte, Int64BytesLen)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, HashData(append(data, timestampBytes)...), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

func SignPSSWithTimestampAndUserID(privateKey *rsa.PrivateKey, userID, timestamp int64, data ...[]byte) ([]byte, error) {
	userIDBytes := make([]byte, Int64BytesLen)
	binary.LittleEndian.PutUint64(userIDBytes, uint64(userID))
	return SignPSSWithTimestamp(privateKey, timestamp, HashData(append(data, userIDBytes)...))
}

func Decrypt(privateKey *rsa.PrivateKey, encryptedAESKey, encryptedData []byte) (data []byte, err error) {
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedAESKey, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA key decryption failed: %w", err)
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short for nonce")
	}
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	data, err = gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM decryption failed: %w", err)
	}
	return data, nil
}

func GenerateTestRSAKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey
}
