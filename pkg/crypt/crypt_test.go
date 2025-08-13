package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHashData(t *testing.T) {
	data1 := []byte("data1")
	data2 := []byte("data2")
	hash := HashData(data1, data2)
	hasher := sha256.New()
	hasher.Write(data1)
	hasher.Write(data2)
	expectedHash := hasher.Sum(nil)
	assert.Equal(t, expectedHash, hash, "Hash should match concatenated SHA256 hash")
}

func TestVerifyPSSWithTimestamp(t *testing.T) {
	privateKey, publicKey := GenerateTestRSAKeyPair(t)
	data := []byte("test data")
	timestamp := time.Now().Unix()

	signature, err := SignPSSWithTimestamp(privateKey, timestamp, data)
	assert.NoError(t, err)

	result := VerifyPSSWithTimestamp(publicKey, signature, timestamp, data)
	assert.True(t, result, "Signature verification should succeed")

	// Test with wrong timestamp
	result = VerifyPSSWithTimestamp(publicKey, signature, timestamp+1, data)
	assert.False(t, result, "Signature verification should fail with wrong timestamp")

	// Test with wrong data
	result = VerifyPSSWithTimestamp(publicKey, signature, timestamp, []byte("wrong data"))
	assert.False(t, result, "Signature verification should fail with wrong data")
}

func TestVerifyPSSWithTimestampAndUserID(t *testing.T) {
	privateKey, publicKey := GenerateTestRSAKeyPair(t)
	data := []byte("test data")
	userID := int64(123)
	timestamp := time.Now().Unix()

	signature, err := SignPSSWithTimestampAndUserID(privateKey, userID, timestamp, data)
	assert.NoError(t, err)

	result := VerifyPSSWithTimestampAndUserID(publicKey, signature, userID, timestamp, data)
	assert.True(t, result, "Signature verification with userID should succeed")

	// Test with wrong userID
	result = VerifyPSSWithTimestampAndUserID(publicKey, signature, userID+1, timestamp, data)
	assert.False(t, result, "Signature verification should fail with wrong userID")

	// Test with wrong timestamp
	result = VerifyPSSWithTimestampAndUserID(publicKey, signature, userID, timestamp+1, data)
	assert.False(t, result, "Signature verification should fail with wrong timestamp")
}

func TestSignPSSWithTimestamp(t *testing.T) {
	privateKey, _ := GenerateTestRSAKeyPair(t)
	data := []byte("test data")
	timestamp := time.Now().Unix()

	signature, err := SignPSSWithTimestamp(privateKey, timestamp, data)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature, "Signature should not be empty")

	// Verify the signature using the corresponding public key
	publicKey := &privateKey.PublicKey
	result := VerifyPSSWithTimestamp(publicKey, signature, timestamp, data)
	assert.True(t, result, "Generated signature should verify correctly")
}

func TestSignPSSWithTimestampAndUserID(t *testing.T) {
	privateKey, _ := GenerateTestRSAKeyPair(t)
	data := []byte("test data")
	userID := int64(123)
	timestamp := time.Now().Unix()

	signature, err := SignPSSWithTimestampAndUserID(privateKey, userID, timestamp, data)
	assert.NoError(t, err)
	assert.NotEmpty(t, signature, "Signature should not be empty")

	// Verify the signature using the corresponding public key
	publicKey := &privateKey.PublicKey
	result := VerifyPSSWithTimestampAndUserID(publicKey, signature, userID, timestamp, data)
	assert.True(t, result, "Generated signature with userID should verify correctly")
}

func TestDecrypt(t *testing.T) {
	privateKey, publicKey := GenerateTestRSAKeyPair(t)
	data := []byte("test data")

	// Generate AES key
	aesKey := make([]byte, AESKeySize)
	_, err := rand.Read(aesKey)
	assert.NoError(t, err)

	// Encrypt AES key with RSA
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	assert.NoError(t, err)

	// Encrypt data with AES-GCM
	block, err := aes.NewCipher(aesKey)
	assert.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	assert.NoError(t, err)
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	assert.NoError(t, err)
	encryptedData := gcm.Seal(nonce, nonce, data, nil)

	// Decrypt
	decryptedData, err := Decrypt(privateKey, encryptedAESKey, encryptedData)
	assert.NoError(t, err)
	assert.Equal(t, data, decryptedData, "Decrypted data should match original")

	// Test with invalid AES key
	_, err = Decrypt(privateKey, []byte("invalid"), encryptedData)
	assert.Error(t, err, "Should error on invalid AES key")

	// Test with invalid encrypted data
	_, err = Decrypt(privateKey, encryptedAESKey, []byte("short"))
	assert.Error(t, err, "Should error on invalid encrypted data")
}
