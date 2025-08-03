package app

import (
	"fmt"
)

func (s *App) GetPrivateKeyPath() string {
	return s.cfg.PrivateKeyPath
}

func (s *App) SetPrivateKeyPath(privateKeyPath string) error {
	err := s.auth.SetPrivateKeyFromPath(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to set private key: %w", err)
	}
	s.auth.SetUserID(0)
	s.cfg.SetPrivateKeyPath(privateKeyPath)
	return nil
}

func (s *App) IsLogged() bool {
	return s.auth.UserID > 0
}

func (s *App) IsOnline() bool {
	return s.auth.UserID > 0
}

func (s *App) Fingerprint() string {
	return s.auth.Fingerprint
}
