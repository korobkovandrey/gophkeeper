package app

import (
	"fmt"
	"gophkeeper/internal/client/model"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"
)

func (a *App) makeSignAuth(id model.ID, data ...[]byte) (auth *proto.Auth, err error) {
	auth = &proto.Auth{
		Id:        string(id),
		UserId:    a.key.UserID,
		Timestamp: a.time.ServerNow().Unix(),
	}
	auth.Signature, err = crypt.SignPSSWithTimestampAndUserID(a.key.PrivateKey, auth.UserId, auth.Timestamp,
		append(data, []byte(auth.Id), a.key.PublicKeyBytes)...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}
	return auth, nil
}
