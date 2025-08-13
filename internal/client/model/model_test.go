package model

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"gophkeeper/pkg/crypt"

	"github.com/stretchr/testify/assert"
)

func TestSecretClone(t *testing.T) {
	original := &Secret{
		ID:          "id1",
		NewID:       "newID1",
		Crypt:       []byte{1, 2, 3},
		Meta:        []byte{4, 5, 6},
		Data:        []byte{7, 8, 9},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Status:      StatusNew,
		Type:        TypeText,
		DecryptMeta: Meta{{Key: "key1", Val: "val1"}},
	}
	assert.Equal(t, original, original.Clone())
}

func TestNewSecret(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	publicKey := &privateKey.PublicKey

	tests := []struct {
		name        string
		id          ID
		newID       ID
		typ         Type
		meta        Meta
		data        []byte
		publicKey   *rsa.PublicKey
		expectError bool
	}{
		{
			name:      "Valid text secret",
			id:        "id1",
			newID:     "newID1",
			typ:       TypeText,
			meta:      Meta{{Key: "key1", Val: "val1"}},
			data:      []byte("test data"),
			publicKey: publicKey,
		},
		{
			name:        "Invalid public key",
			id:          "id2",
			newID:       "newID2",
			typ:         TypeCard,
			meta:        Meta{{Key: "key2", Val: "val2"}},
			data:        []byte("card data"),
			publicKey:   &rsa.PublicKey{}, // Invalid key
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := NewSecret(tt.id, tt.newID, tt.typ, tt.meta, tt.data, tt.publicKey)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.id, secret.ID)
				assert.Equal(t, tt.newID, secret.NewID)
				assert.Equal(t, tt.typ, secret.Type)
				assert.Equal(t, tt.meta, secret.DecryptMeta)
				assert.Equal(t, StatusNew, secret.Status)
				assert.NotEmpty(t, secret.Crypt)
				assert.NotEmpty(t, secret.Meta)
				assert.NotEmpty(t, secret.Data)
			}
		})
	}
}

func TestMakeSecret(t *testing.T) {
	privateKey, publicKey := crypt.GenerateTestRSAKeyPair(t)

	// Prepare test data
	meta := Meta{{Key: "key1", Val: "val1"}}
	data := []byte("test data")
	secret, err := NewSecret("id1", "newID1", TypeText, meta, data, publicKey)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		newID       string
		cryptBytes  []byte
		meta        []byte
		data        []byte
		createdAt   time.Time
		updatedAt   time.Time
		privateKey  *rsa.PrivateKey
		expectError bool
	}{
		{
			name:       "Valid secret",
			id:         "id1",
			newID:      "newID1",
			cryptBytes: secret.Crypt,
			meta:       secret.Meta,
			data:       secret.Data,
			createdAt:  secret.CreatedAt,
			updatedAt:  secret.UpdatedAt,
			privateKey: privateKey,
		},
		{
			name:        "Invalid private key",
			id:          "id2",
			newID:       "newID2",
			cryptBytes:  secret.Crypt,
			meta:        secret.Meta,
			data:        secret.Data,
			createdAt:   secret.CreatedAt,
			updatedAt:   secret.UpdatedAt,
			privateKey:  &rsa.PrivateKey{}, // Invalid key
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := MakeSecret(tt.id, tt.newID, tt.cryptBytes, tt.meta, tt.data, tt.createdAt, tt.updatedAt, tt.privateKey)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, ID(tt.id), secret.ID)
				assert.Equal(t, ID(tt.newID), secret.NewID)
				assert.Equal(t, StatusSynced, secret.Status)
				assert.Equal(t, TypeText, secret.Type)
				assert.Equal(t, meta, secret.DecryptMeta)
			}
		})
	}
}
