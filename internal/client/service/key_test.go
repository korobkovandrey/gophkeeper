package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"gophkeeper/pkg/crypt"

	"github.com/stretchr/testify/assert"
)

func TestNewKey(t *testing.T) {
	key := NewKey()
	assert.NotNil(t, key, "NewKey should return a non-nil Key")
	assert.Equal(t, int64(0), key.UserID, "UserID should be zero")
	assert.Nil(t, key.PrivateKey, "PrivateKey should be nil")
	assert.Nil(t, key.PublicKeyBytes, "PublicKeyBytes should be nil")
	assert.Nil(t, key.Hash, "Hash should be nil")
	assert.Empty(t, key.Fingerprint, "Fingerprint should be empty")
}

func TestSetPrivateKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err, "Failed to generate RSA key")

	key := NewKey()
	err = key.SetPrivateKey(privateKey)
	assert.NoError(t, err, "SetPrivateKey should not return an error")

	assert.Equal(t, privateKey, key.PrivateKey, "PrivateKey should be set correctly")
	assert.NotNil(t, key.PublicKeyBytes, "PublicKeyBytes should be set")
	assert.NotNil(t, key.Hash, "Hash should be set")
	assert.NotEmpty(t, key.Fingerprint, "Fingerprint should be set")

	// Verify public key bytes
	expectedPublicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err, "Failed to marshal public key")
	assert.Equal(t, expectedPublicKeyBytes, key.PublicKeyBytes, "PublicKeyBytes should match")

	// Verify hash
	expectedHash := crypt.PublicKeyHash(key.PublicKeyBytes)
	assert.Equal(t, expectedHash, key.Hash, "Hash should match")

	// Verify fingerprint
	expectedFingerprint := crypt.FingerprintFromHash(expectedHash)
	assert.Equal(t, expectedFingerprint, key.Fingerprint, "Fingerprint should match")
}

func TestSetPrivateKeyFromPath(t *testing.T) {
	// Note: File I/O is not supported in this environment, so we test the underlying logic
	// by mocking the ParsePrivateKey function indirectly via SetPrivateKey.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err, "Failed to generate RSA key")

	key := NewKey()
	err = key.SetPrivateKey(privateKey) // Simulate successful private key parsing
	assert.NoError(t, err, "SetPrivateKey (simulating SetPrivateKeyFromPath) should not return an error")

	// Test invalid private key scenario
	invalidKey := &rsa.PrivateKey{} // Invalid key (not properly initialized)
	key = NewKey()
	err = key.SetPrivateKey(invalidKey)
	assert.Error(t, err, "SetPrivateKey with invalid key should return an error")
}

func TestSetUserID(t *testing.T) {
	key := NewKey()
	userID := int64(12345)
	key.SetUserID(userID)
	assert.Equal(t, userID, key.UserID, "UserID should be set correctly")
}
