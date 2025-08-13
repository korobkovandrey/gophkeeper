package service

import (
	"crypto/rand"
	"crypto/rsa"
	"gophkeeper/internal/client/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupStorage(t *testing.T) (*Storage, *Key, *MemStore) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err, "Failed to generate RSA key")
	key := NewKey()
	err = key.SetPrivateKey(privateKey)
	assert.NoError(t, err, "Failed to set private key")
	memStore := NewMemStore()
	return NewStorage(key, memStore), key, memStore
}

func TestNewStorage(t *testing.T) {
	key := NewKey()
	memStore := NewMemStore()
	storage := NewStorage(key, memStore)
	assert.NotNil(t, storage, "NewStorage should return a non-nil Storage")
	assert.Equal(t, key, storage.key, "Key should be set correctly")
	assert.Equal(t, memStore, storage.m, "MemStore should be set correctly")
}

func TestStorageList(t *testing.T) {
	storage, _, memStore := setupStorage(t)
	secret1 := &model.Secret{
		ID:          "1",
		UpdatedAt:   model.NowTime(),
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}
	secret2 := &model.Secret{
		ID:          "2",
		UpdatedAt:   model.NowTime().Add(-time.Second),
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}
	memStore.store["1"] = secret1
	memStore.store["2"] = secret2

	result := storage.List()
	assert.Len(t, result, 2, "List should return all secrets")
	assert.Equal(t, secret1, result[0], "First secret should be most recently updated")
	assert.Equal(t, secret2, result[1], "Second secret should be less recently updated")
}

func TestStorageSaveText(t *testing.T) {
	storage, _, memStore := setupStorage(t)
	id := model.ID("1")
	newID := model.ID("1")
	meta := model.Meta{{Key: "desc", Val: "text data"}}
	text := "sample text"

	err := storage.SaveText(id, newID, meta, text)
	assert.NoError(t, err, "SaveText should not return an error")
	secret := memStore.store[id]
	assert.NotNil(t, secret, "Secret should be stored")
	assert.Equal(t, model.TypeText, secret.Type, "Secret type should be TypeText")
	assert.Equal(t, meta, secret.DecryptMeta, "Meta should match")
	dataText, _, err := storage.DataText(*secret)
	assert.NoError(t, err, "DataText should decrypt correctly")
	assert.Equal(t, model.DataText{Text: text}, dataText, "Decrypted text should match")
}

func TestStorageSaveLoginPass(t *testing.T) {
	storage, _, memStore := setupStorage(t)
	id := model.ID("1")
	newID := model.ID("1")
	meta := model.Meta{{Key: "desc", Val: "login data"}}
	login, pass := "user", "pass123"

	err := storage.SaveLoginPass(id, newID, meta, login, pass)
	assert.NoError(t, err, "SaveLoginPass should not return an error")
	secret := memStore.store[id]
	assert.NotNil(t, secret, "Secret should be stored")
	assert.Equal(t, model.TypeLoginPass, secret.Type, "Secret type should be TypeLoginPass")
	assert.Equal(t, meta, secret.DecryptMeta, "Meta should match")
	dataLoginPass, _, err := storage.DataLoginPass(*secret)
	assert.NoError(t, err, "DataLoginPass should decrypt correctly")
	assert.Equal(t, model.DataLoginPass{Login: login, Pass: pass}, dataLoginPass, "Decrypted login/pass should match")
}

func TestStorageSaveCard(t *testing.T) {
	storage, _, memStore := setupStorage(t)
	id := model.ID("1")
	newID := model.ID("1")
	meta := model.Meta{{Key: "desc", Val: "card data"}}
	ccn, expire, cvv := "1234567890123456", "12/25", "123"

	err := storage.SaveCard(id, newID, meta, ccn, expire, cvv)
	assert.NoError(t, err, "SaveCard should not return an error")
	secret := memStore.store[id]
	assert.NotNil(t, secret, "Secret should be stored")
	assert.Equal(t, model.TypeCard, secret.Type, "Secret type should be TypeCard")
	assert.Equal(t, meta, secret.DecryptMeta, "Meta should match")
	dataCard, _, err := storage.DataCard(*secret)
	assert.NoError(t, err, "DataCard should decrypt correctly")
	assert.Equal(t, model.DataCard{CCN: ccn, Expire: expire, CVV: cvv}, dataCard, "Decrypted card data should match")
}

func TestStorageDataText(t *testing.T) {
	storage, key, memStore := setupStorage(t)
	id := model.ID("1")
	newID := model.ID("1")
	meta := model.Meta{{Key: "desc", Val: "text data"}}
	text := "sample text"

	secret, err := model.NewSecret(id, newID, model.TypeText, meta, []byte(`{"Text":"sample text"}`), &key.PrivateKey.PublicKey)
	assert.NoError(t, err, "Failed to create secret")
	memStore.store[id] = secret

	dataText, decryptedMeta, err := storage.DataText(*secret)
	assert.NoError(t, err, "DataText should not return an error")
	assert.Equal(t, model.DataText{Text: text}, dataText, "Decrypted text should match")
	assert.Equal(t, meta, decryptedMeta, "Decrypted meta should match")

	// Test decryption failure
	secret.Crypt = []byte("invalid")
	_, _, err = storage.DataText(*secret)
	assert.Error(t, err, "DataText should return error for invalid decryption")
}

func TestStorageDataLoginPass(t *testing.T) {
	storage, key, memStore := setupStorage(t)
	id := model.ID("1")
	newID := model.ID("1")
	meta := model.Meta{{Key: "desc", Val: "login data"}}
	login, pass := "user", "pass123"

	secret, err := model.NewSecret(id, newID, model.TypeLoginPass, meta, []byte(`{"Login":"user","Pass":"pass123"}`), &key.PrivateKey.PublicKey)
	assert.NoError(t, err, "Failed to create secret")
	memStore.store[id] = secret

	dataLoginPass, decryptedMeta, err := storage.DataLoginPass(*secret)
	assert.NoError(t, err, "DataLoginPass should not return an error")
	assert.Equal(t, model.DataLoginPass{Login: login, Pass: pass}, dataLoginPass, "Decrypted login/pass should match")
	assert.Equal(t, meta, decryptedMeta, "Decrypted meta should match")

	// Test invalid JSON
	secret.Data = []byte("invalid json")
	_, _, err = storage.DataLoginPass(*secret)
	assert.Error(t, err, "DataLoginPass should return error for invalid JSON")
}

func TestStorageDataCard(t *testing.T) {
	storage, key, memStore := setupStorage(t)
	id := model.ID("1")
	newID := model.ID("1")
	meta := model.Meta{{Key: "desc", Val: "card data"}}
	ccn, expire, cvv := "1234567890123456", "12/25", "123"

	secret, err := model.NewSecret(id, newID, model.TypeCard, meta, []byte(`{"CCN":"1234567890123456","Expire":"12/25","CVV":"123"}`), &key.PrivateKey.PublicKey)
	assert.NoError(t, err, "Failed to create secret")
	memStore.store[id] = secret

	dataCard, decryptedMeta, err := storage.DataCard(*secret)
	assert.NoError(t, err, "DataCard should not return an error")
	assert.Equal(t, model.DataCard{CCN: ccn, Expire: expire, CVV: cvv}, dataCard, "Decrypted card data should match")
	assert.Equal(t, meta, decryptedMeta, "Decrypted meta should match")

	// Test decryption failure
	secret.Crypt = []byte("invalid")
	_, _, err = storage.DataCard(*secret)
	assert.Error(t, err, "DataCard should return error for invalid decryption")
}

func TestStorageGet(t *testing.T) {
	storage, _, memStore := setupStorage(t)
	secret := &model.Secret{
		ID:          "1",
		UpdatedAt:   model.NowTime(),
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}
	memStore.store["1"] = secret

	result, err := storage.Get("1")
	assert.NoError(t, err, "Get should not return an error for existing secret")
	assert.Equal(t, secret, result, "Get should return correct secret")

	result, err = storage.Get("2")
	assert.Error(t, err, "Get should return error for non-existent secret")
	assert.Nil(t, result, "Get should return nil for non-existent secret")
	assert.Contains(t, err.Error(), "secret not found", "Error should mention secret not found")
}

func TestStorageDelete(t *testing.T) {
	storage, _, memStore := setupStorage(t)
	memStore.store["1"] = &model.Secret{
		ID:          "1",
		UpdatedAt:   model.NowTime(),
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
		Status:      model.StatusNew,
	}
	memStore.store["2"] = &model.Secret{
		ID:          "2",
		UpdatedAt:   model.NowTime(),
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
		Status:      model.StatusSynced,
	}
	storage.Delete("1")
	storage.Delete("2")
	assert.Nil(t, memStore.store["1"], "Delete should remove secret from store")
	assert.Equal(t, model.StatusDeleting, memStore.store["2"].Status, "Delete should mark secret as StatusDeleting")
}

func TestStorageClear(t *testing.T) {
	storage, _, memStore := setupStorage(t)
	memStore.store["1"] = &model.Secret{ID: "1"}
	memStore.store["2"] = &model.Secret{ID: "2"}

	storage.Clear()
	assert.Empty(t, memStore.store, "Clear should empty the store")
}
