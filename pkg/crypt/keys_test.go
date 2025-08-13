package crypt

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

func TestPublicKeyHash(t *testing.T) {
	data := []byte("test public key")
	hash := PublicKeyHash(data)
	assert.Len(t, hash, sha256.Size, "Hash length should match SHA256 size")
	expectedHash := sha256.Sum256(data)
	assert.Equal(t, expectedHash[:], hash, "Hash should match SHA256 of input")
}

func TestFingerprintFromHash(t *testing.T) {
	hash := []byte("test hash")
	fingerprint := FingerprintFromHash(hash)
	expectedFingerprint := base64.StdEncoding.EncodeToString(hash)
	assert.Equal(t, expectedFingerprint, fingerprint, "Fingerprint should be base64 encoded hash")
}

func TestFingerprint(t *testing.T) {
	data := []byte("test public key")
	fingerprint := Fingerprint(data)
	expectedHash := sha256.Sum256(data)
	expectedFingerprint := base64.StdEncoding.EncodeToString(expectedHash[:])
	assert.Equal(t, expectedFingerprint, fingerprint, "Fingerprint should match base64 encoded SHA256 hash")
}

func TestRSAPublicKeyFromBytes(t *testing.T) {
	_, publicKey := GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	assert.NoError(t, err)

	parsedKey, err := RSAPublicKeyFromBytes(publicKeyBytes)
	assert.NoError(t, err)
	assert.Equal(t, publicKey, parsedKey, "Parsed public key should match original")

	// Test invalid bytes
	_, err = RSAPublicKeyFromBytes([]byte("invalid"))
	assert.Error(t, err, "Should error on invalid public key bytes")

	// Test non-RSA key (use a different type, e.g., empty bytes for invalid format)
	_, err = RSAPublicKeyFromBytes([]byte{})
	assert.Error(t, err, "Should error on non-RSA public key")
}

func TestEncodePublicKey(t *testing.T) {
	publicKeyBytes := []byte("test public key")
	encoded := EncodePublicKey(publicKeyBytes)
	expected := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))
	assert.Equal(t, expected, encoded, "Encoded public key should match PEM format")
}

func TestDecodePublicKey(t *testing.T) {
	publicKeyBytes := []byte("test public key")
	pemStr := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	}))

	decodedBytes, err := DecodePublicKey(pemStr)
	assert.NoError(t, err)
	assert.Equal(t, publicKeyBytes, decodedBytes, "Decoded public key should match original bytes")

	// Test invalid PEM
	_, err = DecodePublicKey("invalid PEM")
	assert.Error(t, err, "Should error on invalid PEM format")

	// Test wrong block type
	wrongBlock := string(pem.EncodeToMemory(&pem.Block{
		Type:  "WRONG TYPE",
		Bytes: publicKeyBytes,
	}))
	_, err = DecodePublicKey(wrongBlock)
	assert.Error(t, err, "Should error on wrong PEM block type")
}

func TestParsePrivateKey(t *testing.T) {
	privateKey, _ := GenerateTestRSAKeyPair(t)
	sshPrivateKey, err := ssh.MarshalPrivateKey(privateKey, "test")
	assert.NoError(t, err)

	// Write to a temporary file
	tmpFile, err := os.CreateTemp("", "testkey")
	assert.NoError(t, err)
	tmpFileName := tmpFile.Name()
	defer func() {
		assert.NoError(t, os.Remove(tmpFileName))
	}()
	_, err = tmpFile.Write(pem.EncodeToMemory(sshPrivateKey))
	assert.NoError(t, err)
	assert.NoError(t, tmpFile.Close())

	parsedKey, err := ParsePrivateKey(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, privateKey, parsedKey, "Parsed private key should match original")

	// Test invalid file path
	_, err = ParsePrivateKey("nonexistent")
	assert.Error(t, err, "Should error on nonexistent file")

	// Test invalid key format
	tmpFile, err = os.CreateTemp("", "invalidkey")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(tmpFile.Name()))
	}()
	_, err = tmpFile.Write([]byte("invalid key"))
	assert.NoError(t, err)
	assert.NoError(t, tmpFile.Close())
	_, err = ParsePrivateKey(tmpFile.Name())
	assert.Error(t, err, "Should error on invalid key format")
}
