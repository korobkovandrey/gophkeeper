package tuiadapter

import (
	"fmt"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/service"
)

type Key struct {
	cfg *config.Config
	key *service.Key
}

func NewKey(cfg *config.Config, key *service.Key) *Key {
	return &Key{
		cfg: cfg,
		key: key,
	}
}

func (s *Key) GetPrivateKeyPath() string {
	return s.cfg.PrivateKeyPath
}

func (s *Key) SetPrivateKeyPath(privateKeyPath string) error {
	err := s.key.SetPrivateKeyFromPath(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to set private key: %w", err)
	}
	s.key.SetUserID(0)
	s.cfg.SetPrivateKeyPath(privateKeyPath)
	return nil
}

func (s *Key) IsLogged() bool {
	return s.key.UserID > 0
}

func (s *Key) Fingerprint() string {
	return s.key.Fingerprint
}
