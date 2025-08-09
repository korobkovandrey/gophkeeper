package service

import (
	"errors"
	"gophkeeper/internal/client/model"
	"sort"
	"sync"
	"time"
)

type MemStore struct {
	store  map[model.ID]*model.Secret
	mutex  sync.Mutex
	SyncCh chan struct{}
}

func NewMemStore() *MemStore {
	return &MemStore{
		store:  make(map[model.ID]*model.Secret),
		SyncCh: make(chan struct{}, 1),
	}
}

func (ms *MemStore) Store(secret *model.Secret) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	if existing, ok := ms.store[secret.ID]; ok {
		if existing.Status == model.StatusNew {
			if secret.ID != secret.NewID {
				delete(ms.store, secret.ID)
				secret.ID = secret.NewID
				if existing, ok = ms.store[secret.ID]; ok && existing.Status != model.StatusNew {
					secret.Status = model.StatusUpdating
				}
			}
		} else {
			secret.Status = model.StatusUpdating
		}
	}
	ms.store[secret.ID] = secret
	ms.sync()
	return nil
}

func (ms *MemStore) Synced(secret *model.Secret) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	existing, existsOld := ms.store[secret.ID]
	if existsOld {
		if secret.UpdatedAt.Unix() < existing.UpdatedAt.Unix() {
			return errors.New("conflict")
		}
	}
	if secret.ID != secret.NewID {
		if existingNew, ok := ms.store[secret.NewID]; ok {
			if secret.UpdatedAt.Unix() < existingNew.UpdatedAt.Unix() {
				return errors.New("conflict")
			}
		}
		if existsOld {
			delete(ms.store, secret.ID)
		}
		secret.ID = secret.NewID
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
	ms.sync()
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
	for id := range ms.store {
		secrets = append(secrets, ms.store[id].Clone())
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

func (ms *MemStore) sync() {
	select {
	case ms.SyncCh <- struct{}{}:
	default:
	}
}

func (ms *MemStore) ListForSync() []*model.Secret {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	secrets := make([]*model.Secret, 0, len(ms.store))
	for id := range ms.store {
		if ms.store[id].Status == model.StatusSynced {
			continue
		}
		secrets = append(secrets, ms.store[id].Clone())
	}
	sort.Slice(secrets, func(i, j int) bool {
		return secrets[i].UpdatedAt.Before(secrets[j].UpdatedAt)
	})
	return secrets
}

func (ms *MemStore) Close() {
	close(ms.SyncCh)
}
