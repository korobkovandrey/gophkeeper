package service

import (
	"errors"
	"gophkeeper/internal/client/model"
	"sort"
	"sync"
	"time"
)

type MemStore struct {
	store map[model.ID]*model.Secret
	mutex sync.Mutex
}

func NewMemStore() *MemStore {
	return &MemStore{
		store: make(map[model.ID]*model.Secret),
	}
}

func (ms *MemStore) Store(id model.ID, secret *model.Secret) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	existing, existsOld := ms.store[id]
	if existsOld {
		if existing.UpdatedAt.After(secret.UpdatedAt) {
			return errors.New("conflict")
		}
		if existing.Status != model.StatusNew {
			secret.Status = model.StatusUpdating
		}
	}
	if id != secret.ID {
		var exists bool
		existing, exists = ms.store[secret.ID]
		if exists {
			if existing.UpdatedAt.After(secret.UpdatedAt) {
				return errors.New("conflict")
			}
			if existing.Status != model.StatusNew {
				secret.Status = model.StatusUpdating
			}
		}
		if existsOld {
			delete(ms.store, id)
		}
	}
	ms.store[secret.ID] = secret
	return nil
}

func (ms *MemStore) Get(id model.ID) *model.Secret {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	secret, exists := ms.store[id]
	if !exists {
		return nil
	}
	return secret.Clone()
}

func (ms *MemStore) Deleting(id model.ID) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	if _, ok := ms.store[id]; !ok {
		return
	}
	if ms.store[id].Status == model.StatusNew {
		delete(ms.store, id)
		return
	}
	ms.store[id].Status = model.StatusDeleting
	ms.store[id].UpdatedAt = time.Now()
}

func (ms *MemStore) Delete(id model.ID) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	delete(ms.store, id)
}

func (ms *MemStore) List() []*model.Secret {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	secrets := make([]*model.Secret, 0, len(ms.store))
	for _, secret := range ms.store {
		secrets = append(secrets, secret.Clone())
	}
	sort.Slice(secrets, func(i, j int) bool {
		return secrets[i].UpdatedAt.After(secrets[j].UpdatedAt)
	})
	return secrets
}

func (ms *MemStore) Clear() {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.store = make(map[model.ID]*model.Secret)
}
