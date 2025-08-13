package service

import (
	"crypto/x509"
	"fmt"
	"testing"
	"time"

	"gophkeeper/internal/server/infra/db/query"
	"gophkeeper/pkg/crypt"

	"github.com/stretchr/testify/assert"
)

func TestNewUserFromQueryUser(t *testing.T) {
	_, publicKey := crypt.GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	assert.NoError(t, err)
	encodedPublicKey := crypt.EncodePublicKey(publicKeyBytes)
	fingerprint := crypt.Fingerprint(publicKeyBytes)
	fixedTime := time.Now().Truncate(time.Second)

	tests := []struct {
		name         string
		input        query.User
		expectedUser *User
		expectedErr  error
	}{
		{
			name: "Successful conversion",
			input: query.User{
				ID:          1,
				Fingerprint: fingerprint,
				PublicKey:   encodedPublicKey,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
			expectedUser: &User{
				ID:             1,
				Fingerprint:    fingerprint,
				PublicKey:      encodedPublicKey,
				CreatedAt:      fixedTime,
				UpdatedAt:      fixedTime,
				PublicKeyBytes: publicKeyBytes,
			},
			expectedErr: nil,
		},
		{
			name: "Invalid public key",
			input: query.User{
				ID:          1,
				Fingerprint: fingerprint,
				PublicKey:   "invalid-public-key", // Invalid PEM format
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
			expectedUser: nil,
			expectedErr:  fmt.Errorf("failed to decode public key: invalid public key format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := newUserFromQueryUser(tt.input)
			assert.Equal(t, tt.expectedUser, user, "expected user %v, got %v", tt.expectedUser, user)
			if tt.expectedErr == nil {
				assert.NoError(t, err, "expected no error")
			} else {
				assert.Error(t, err, "expected an error")
				assert.EqualError(t, err, tt.expectedErr.Error(), "expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestNewSecretFromQuerySecret(t *testing.T) {
	fixedTime := time.Now().Truncate(time.Second)
	input := query.Secret{
		ID:        "secret-123",
		UserID:    1,
		Crypt:     []byte("encrypted-data"),
		Meta:      []byte("metadata"),
		Data:      []byte("secret-data"),
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}
	expected := &Secret{
		ID:        "secret-123",
		UserID:    1,
		Crypt:     []byte("encrypted-data"),
		Meta:      []byte("metadata"),
		Data:      []byte("secret-data"),
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}

	user := newSecretFromQuerySecret(input)
	assert.Equal(t, expected, user, "expected secret %v, got %v", expected, user)
}
