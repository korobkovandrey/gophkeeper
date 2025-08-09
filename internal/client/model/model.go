package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"gophkeeper/pkg/crypt"
	"time"
)

type ID string
type Status int
type Type byte

const (
	StatusNew Status = iota
	StatusUpdating
	StatusSynced
	StatusDeleting
)

const (
	TypeUnknown Type = iota
	TypeText
	TypeLoginPass
	TypeCard
)

type Secret struct {
	ID          ID
	NewID       ID
	Crypt       []byte
	Meta        []byte
	Data        []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Status      Status
	Type        Type
	DecryptMeta Meta
}

func (s *Secret) Clone() *Secret {
	secret := &Secret{
		ID:          s.ID,
		NewID:       s.NewID,
		Crypt:       make([]byte, len(s.Crypt)),
		Meta:        make([]byte, len(s.Meta)),
		Data:        make([]byte, len(s.Data)),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		Status:      s.Status,
		Type:        s.Type,
		DecryptMeta: make(Meta, len(s.DecryptMeta)),
	}
	copy(secret.Crypt, s.Crypt)
	copy(secret.Meta, s.Meta)
	copy(secret.Data, s.Data)
	copy(secret.DecryptMeta, s.DecryptMeta)
	return secret
}

func NewSecret(id, newID ID, typ Type, meta Meta, data []byte, publicKey *rsa.PublicKey) (*Secret, error) {
	s := &Secret{
		ID:          id,
		NewID:       newID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Status:      StatusNew,
		Type:        typ,
		DecryptMeta: meta,
	}
	metaBytes, err := meta.MarshalWithType(typ)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta with type: %w", err)
	}
	aesKey := make([]byte, crypt.AESKeySize)
	_, err = rand.Read(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}
	s.Crypt, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("RSA encryption failed: %w", err)
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}
	s.Meta = gcm.Seal(nonce, nonce, metaBytes, nil)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}
	s.Data = gcm.Seal(nonce, nonce, data, nil)
	return s, nil
}
