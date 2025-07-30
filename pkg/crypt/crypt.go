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
)

func HashData(data ...[]byte) []byte {
	hasher := sha256.New()
	for _, d := range data {
		hasher.Write(d)
	}
	return hasher.Sum(nil)
}

func VerifyPSS(rsaPublicKey *rsa.PublicKey, signature []byte, data ...[]byte) bool {
	return rsa.VerifyPSS(rsaPublicKey, crypto.SHA256, HashData(data...), signature, nil) == nil
}

func SignPSS(privateKey *rsa.PrivateKey, data ...[]byte) ([]byte, error) {
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, HashData(data...), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

func VerifyPSSWithTimestamp(rsaPublicKey *rsa.PublicKey, signature []byte, timestamp int64, data ...[]byte) bool {
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	return rsa.VerifyPSS(rsaPublicKey, crypto.SHA256, HashData(append(data, timestampBytes)...), signature, nil) == nil
}

func VerifyPSSWithTimestampAndUserID(rsaPublicKey *rsa.PublicKey, signature []byte, userID, timestamp int64, data ...[]byte) bool {
	userIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(userIDBytes, uint64(userID))
	return VerifyPSSWithTimestamp(rsaPublicKey, signature, timestamp, HashData(append(data, userIDBytes)...))
}

func SignPSSWithTimestamp(privateKey *rsa.PrivateKey, timestamp int64, data ...[]byte) ([]byte, error) {
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, uint64(timestamp))
	signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, HashData(append(data, timestampBytes)...), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

func SignPSSWithTimestampAndUserID(privateKey *rsa.PrivateKey, userID, timestamp int64, data ...[]byte) ([]byte, error) {
	userIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(userIDBytes, uint64(userID))
	return SignPSSWithTimestamp(privateKey, timestamp, HashData(append(data, userIDBytes)...))
}

const oaep256Size = 66

func IsWithAESKeyEncrypting(publicKey *rsa.PublicKey, data []byte) bool {
	return len(data) > publicKey.Size()-oaep256Size
}

func Encrypt(publicKey *rsa.PublicKey, data []byte) (encryptedData []byte, err error) {
	if IsWithAESKeyEncrypting(publicKey, data) {
		return nil, fmt.Errorf("data too long for RSA encryption")
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
}

func EncryptWithAESKey(publicKey *rsa.PublicKey, data []byte) (encryptedAESKey, encryptedData []byte, err error) {
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, nil, fmt.Errorf("failed to generate AES key: %w", err)
	}
	encryptedAESKey, err = Encrypt(publicKey, aesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("RSA encryption failed: %w", err)
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %v", err)
	}
	encryptedData = gcm.Seal(nonce, nonce, data, nil)

	return encryptedAESKey, encryptedData, nil
}

func Decrypt(privateKey *rsa.PrivateKey, encryptedAESKey, encryptedData []byte) (data []byte, err error) {
	if len(encryptedAESKey) == 0 {
		data, err = rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedData, nil)
		if err != nil {
			return nil, fmt.Errorf("RSA decryption failed: %w", err)
		}
		return data, nil
	}
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
