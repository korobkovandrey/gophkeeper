package service

import (
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/model"
	"gophkeeper/pkg/crypt"
)

type Storage struct {
	key *Key
	m   *MemStore
}

func NewStorage(key *Key, m *MemStore) *Storage {
	return &Storage{
		key: key,
		m:   m,
	}
}

func (s *Storage) List() []*model.Secret {
	return s.m.List()
}

func (s *Storage) save(id, newID model.ID, meta model.Meta, dataModel any) error {
	data, err := json.Marshal(dataModel)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	var typ model.Type
	switch dataModel.(type) {
	case model.DataText:
		typ = model.TypeText
	case model.DataLoginPass:
		typ = model.TypeLoginPass
	case model.DataCard:
		typ = model.TypeCard
	default:
		return fmt.Errorf("unknown data type: %T", dataModel)
	}
	secret, err := model.NewSecret(newID, typ, meta, data, &s.key.PrivateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}
	err = s.m.Store(id, secret)
	if err != nil {
		return fmt.Errorf("failed to store secret: %w", err)
	}
	return nil
}

func (s *Storage) SaveText(id, newID model.ID, meta model.Meta, text string) error {
	return s.save(id, newID, meta, model.DataText{Text: text})
}

func (s *Storage) SaveLoginPass(id, newID model.ID, meta model.Meta, login, pass string) error {
	return s.save(id, newID, meta, model.DataLoginPass{Login: login, Pass: pass})
}

func (s *Storage) SaveCard(id, newID model.ID, meta model.Meta, ccn, expire, cvv string) error {
	return s.save(id, newID, meta, model.DataCard{CCN: ccn, Expire: expire, CVV: cvv})
}

func (s *Storage) DataText(id model.ID) (model.DataText, model.Meta, error) {
	return decryptSecret[model.DataText](s, id)
}

func (s *Storage) DataLoginPass(id model.ID) (model.DataLoginPass, model.Meta, error) {
	return decryptSecret[model.DataLoginPass](s, id)
}

func (s *Storage) DataCard(id model.ID) (model.DataCard, model.Meta, error) {
	return decryptSecret[model.DataCard](s, id)
}

func decryptSecret[T any](s *Storage, id model.ID) (T, model.Meta, error) {
	secret := s.m.Get(id)
	var data T
	if secret == nil {
		return data, model.Meta{}, fmt.Errorf("secret not found %v", id)
	}
	decryptData, err := crypt.Decrypt(s.key.PrivateKey, secret.Crypt, secret.Data)
	if err != nil {
		return data, model.Meta{}, fmt.Errorf("failed to decrypt data: %w", err)
	}
	err = json.Unmarshal(decryptData, &data)
	if err != nil {
		return data, model.Meta{}, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return data, secret.DecryptMeta, nil
}

func (s *Storage) Type(id model.ID) (model.Type, error) {
	secret := s.m.Get(id)
	if secret == nil {
		return model.TypeUnknown, fmt.Errorf("secret not found %v", id)
	}
	return secret.Type, nil
}

func (s *Storage) Delete(id model.ID) {
	s.m.Deleting(id)
}

func (s *Storage) Clear() {
	s.m.Clear()
}
