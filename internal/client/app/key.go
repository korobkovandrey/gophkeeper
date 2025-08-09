package app

import (
	"fmt"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/service"
)

type KeyManager struct {
	cfg *config.Config
	key *service.Key
}

func NewKeyManager(cfg *config.Config, key *service.Key) *KeyManager {
	return &KeyManager{
		cfg: cfg,
		key: key,
	}
}

func (s *KeyManager) GetPrivateKeyPath() string {
	return s.cfg.PrivateKeyPath
}

func (s *KeyManager) SetPrivateKeyPath(privateKeyPath string) error {
	err := s.key.SetPrivateKeyFromPath(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to set private key: %w", err)
	}
	s.key.SetUserID(0)
	s.cfg.SetPrivateKeyPath(privateKeyPath)
	return nil
}

func (s *KeyManager) Fingerprint() string {
	return s.key.Fingerprint
}
