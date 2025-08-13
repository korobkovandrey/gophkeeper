package service

import (
	"fmt"
	"gophkeeper/internal/client/model"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStore(t *testing.T) {
	store := NewMemStore()
	assert.NotNil(t, store, "NewMemStore should return a non-nil MemStore")
	assert.NotNil(t, store.store, "Store map should be initialized")
	assert.NotNil(t, store.SyncCh, "SyncCh should be initialized")
	assert.Equal(t, 1, cap(store.SyncCh), "SyncCh should have capacity 1")
}

func TestMemStoreStore(t *testing.T) {
	store := NewMemStore()
	secret := &model.Secret{
		ID:        "1",
		NewID:     "1",
		CreatedAt: model.NowTime(),
		UpdatedAt: model.NowTime(),
		Status:    model.StatusNew,
	}

	// Test storing a new secret
	err := store.Store(secret)
	assert.NoError(t, err, "Store should not return an error for new secret")
	assert.Equal(t, secret, store.store["1"], "Secret should be stored correctly")

	// Test updating existing secret with same ID
	updatedSecret := &model.Secret{
		ID:        "1",
		NewID:     "1",
		UpdatedAt: model.NowTime().Add(time.Second),
		Status:    model.StatusNew,
	}
	err = store.Store(updatedSecret)
	assert.NoError(t, err, "Store should not return an error for updating secret")
	assert.Equal(t, secret.CreatedAt, store.store["1"].CreatedAt, "CreatedAt should be preserved")
	assert.Equal(t, updatedSecret.UpdatedAt, store.store["1"].UpdatedAt, "UpdatedAt should be updated")

	// Test storing with new ID
	newIDSecret := &model.Secret{
		ID:        "1",
		NewID:     "2",
		UpdatedAt: model.NowTime().Add(2 * time.Second),
		Status:    model.StatusNew,
	}
	err = store.Store(newIDSecret)
	assert.NoError(t, err, "Store should not return an error for new ID")
	assert.Nil(t, store.store["1"], "Old ID should be removed")
	assert.Equal(t, newIDSecret, store.store["2"], "Secret should be stored with new ID")

	// Test sync channel
	select {
	case <-store.SyncCh:
		// Expected: sync signal sent
	default:
		t.Fatal("SyncCh should have received a signal")
	}
}

func TestMemStoreSyncStore(t *testing.T) {
	store := NewMemStore()
	secret := &model.Secret{
		ID:        "1",
		NewID:     "1",
		UpdatedAt: model.NowTime(),
		Status:    model.StatusSynced,
	}

	// Test storing new secret
	err := store.SyncStore(secret)
	assert.NoError(t, err, "SyncStore should not return an error for new secret")
	assert.Equal(t, secret, store.store["1"], "Secret should be stored correctly")

	// Test conflict with older timestamp
	olderSecret := &model.Secret{
		ID:        "1",
		NewID:     "1",
		UpdatedAt: secret.UpdatedAt.Add(-time.Second),
		Status:    model.StatusSynced,
	}
	err = store.SyncStore(olderSecret)
	assert.Error(t, err, "SyncStore should return conflict error for older timestamp")
	assert.Contains(t, err.Error(), "conflict", "Error should mention conflict")

	// Test changing ID
	newIDSecret := &model.Secret{
		ID:        "1",
		NewID:     "2",
		UpdatedAt: secret.UpdatedAt.Add(time.Second),
		Status:    model.StatusSynced,
	}
	err = store.SyncStore(newIDSecret)
	assert.NoError(t, err, "SyncStore should not return an error for new ID")
	assert.Nil(t, store.store["1"], "Old ID should be removed")
	assert.Equal(t, newIDSecret, store.store["2"], "Secret should be stored with new ID")
}

func TestMemStoreGet(t *testing.T) {
	store := NewMemStore()
	secret := &model.Secret{
		ID:          "1",
		NewID:       "1",
		UpdatedAt:   model.NowTime(),
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}
	store.store["1"] = secret

	// Test getting existing secret
	result := store.Get("1")
	assert.Equal(t, secret, result, "Get should return correct secret")
	assert.NotSame(t, secret, result, "Get should return a clone")

	// Test getting non-existent secret
	result = store.Get("2")
	assert.Nil(t, result, "Get should return nil for non-existent ID")
}

func TestMemStoreDeleting(t *testing.T) {
	store := NewMemStore()
	secret := &model.Secret{
		ID:          "1",
		NewID:       "1",
		UpdatedAt:   model.NowTime(),
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}

	// Test deleting non-existent secret
	store.Deleting("1")
	assert.Nil(t, store.store["1"], "Deleting non-existent secret should do nothing")

	// Test deleting StatusNew secret
	secret.Status = model.StatusNew
	store.store["1"] = secret
	store.Deleting("1")
	assert.Nil(t, store.store["1"], "StatusNew secret should be deleted")

	// Test deleting non-StatusNew secret
	secret.Status = model.StatusSynced
	store.store["1"] = secret
	store.Deleting("1")
	assert.Equal(t, model.StatusDeleting, store.store["1"].Status, "Status should be set to Deleting")

	// Test sync channel
	select {
	case <-store.SyncCh:
		// Expected: sync signal sent
	default:
		t.Fatal("SyncCh should have received a signal")
	}
}

func TestMemStoreDelete(t *testing.T) {
	store := NewMemStore()
	secret := &model.Secret{
		ID:        "1",
		NewID:     "1",
		UpdatedAt: model.NowTime(),
	}
	store.store["1"] = secret

	// Test deleting with newer timestamp
	err := store.Delete("1", secret.UpdatedAt.Add(time.Second))
	assert.NoError(t, err, "Delete should not return an error for newer timestamp")
	assert.Nil(t, store.store["1"], "Secret should be deleted")

	// Test conflict with older timestamp
	store.store["1"] = secret
	err = store.Delete("1", secret.UpdatedAt.Add(-time.Second))
	assert.Error(t, err, "Delete should return conflict error for older timestamp")
	assert.Contains(t, err.Error(), "conflict", "Error should mention conflict")
	assert.NotNil(t, store.store["1"], "Secret should not be deleted")

	// Test deleting non-existent secret
	err = store.Delete("2", model.NowTime())
	assert.NoError(t, err, "Delete should not return an error for non-existent secret")
}

func TestMemStoreList(t *testing.T) {
	store := NewMemStore()
	secret1 := &model.Secret{
		ID:          "1",
		NewID:       "1",
		UpdatedAt:   model.NowTime(),
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}
	secret2 := &model.Secret{
		ID:          "2",
		NewID:       "2",
		UpdatedAt:   secret1.UpdatedAt.Add(-time.Second),
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}
	store.store["1"] = secret1
	store.store["2"] = secret2

	result := store.List()
	assert.Len(t, result, 2, "List should return all secrets")
	assert.Equal(t, secret1, result[0], "First secret should be most recently updated")
	assert.Equal(t, secret2, result[1], "Second secret should be less recently updated")
	assert.NotSame(t, secret1, result[0], "Secrets should be clones")
	assert.NotSame(t, secret2, result[1], "Secrets should be clones")
}

func TestMemStoreClear(t *testing.T) {
	store := NewMemStore()
	store.store["1"] = &model.Secret{ID: "1"}
	store.store["2"] = &model.Secret{ID: "2"}

	store.Clear()
	assert.Empty(t, store.store, "Clear should empty the store")
}

func TestMemStoreListForSync(t *testing.T) {
	store := NewMemStore()
	secret1 := &model.Secret{
		ID:          "1",
		NewID:       "1",
		UpdatedAt:   model.NowTime(),
		Status:      model.StatusNew,
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}
	secret2 := &model.Secret{
		ID:          "2",
		NewID:       "2",
		UpdatedAt:   secret1.UpdatedAt.Add(time.Second),
		Status:      model.StatusUpdating,
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}
	secret3 := &model.Secret{
		ID:          "3",
		NewID:       "3",
		UpdatedAt:   secret1.UpdatedAt.Add(2 * time.Second),
		Status:      model.StatusSynced,
		Crypt:       make([]byte, 0),
		Meta:        make([]byte, 0),
		Data:        make([]byte, 0),
		DecryptMeta: make(model.Meta, 0),
	}
	store.store["1"] = secret1
	store.store["2"] = secret2
	store.store["3"] = secret3

	result := store.ListForSync()
	assert.Len(t, result, 2, "ListForSync should return only non-Synced secrets")
	assert.Equal(t, secret1, result[0], "First secret should be oldest updated")
	assert.Equal(t, secret2, result[1], "Second secret should be newer")
	assert.NotSame(t, secret1, result[0], "Secrets should be clones")
	assert.NotSame(t, secret2, result[1], "Secrets should be clones")
}

func TestMemStoreClose(t *testing.T) {
	store := NewMemStore()
	store.Close()

	_, ok := <-store.SyncCh
	assert.False(t, ok, "SyncCh should be closed")
}

func TestMemStoreConcurrentAccess(t *testing.T) {
	store := NewMemStore()
	var wg sync.WaitGroup
	numGoroutines := 10
	// Test concurrent Store
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			err := store.Store(&model.Secret{
				ID:        model.ID(fmt.Sprintf("%d", i)),
				NewID:     model.ID(fmt.Sprintf("%d", i)),
				UpdatedAt: model.NowTime(),
			})
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()
	assert.Len(t, store.store, numGoroutines, "All secrets should be stored")

	// Test concurrent Get
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			store.Get(model.ID(fmt.Sprintf("%d", i)))
		}(i)
	}
	wg.Wait()
}
