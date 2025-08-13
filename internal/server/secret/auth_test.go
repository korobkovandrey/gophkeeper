package secret

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"testing"
	"time"

	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stubFindUserByIDFunc creates a stub for FindUserByIDFunc.
type stubFindUserByIDFunc struct {
	User *service.User
	Err  error
}

func (s *stubFindUserByIDFunc) FindUserByID(context.Context, int64) (*service.User, error) {
	return s.User, s.Err
}

func TestServiceServer_findAndAuthUser(t *testing.T) {
	fixedTime := time.Now().Truncate(time.Second).UTC()

	userID := int64(1)
	secretID := "secret-123"
	timestamp := fixedTime.Unix()
	errDB := errors.New("database error")

	privateKey, publicKey := crypt.GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)
	fingerprint := crypt.Fingerprint(publicKeyBytes)
	encodedPublicKey := crypt.EncodePublicKey(publicKeyBytes)
	data := [][]byte{
		[]byte("crypt-data"),
		[]byte("meta-data"),
		[]byte("secret-data"),
	}
	signature, err := crypt.SignPSSWithTimestampAndUserID(privateKey, userID, timestamp, append(data, []byte(secretID), publicKeyBytes)...)
	require.NoError(t, err)
	user := &service.User{
		ID:             userID,
		Fingerprint:    fingerprint,
		PublicKey:      encodedPublicKey,
		PublicKeyBytes: publicKeyBytes,
		CreatedAt:      fixedTime,
		UpdatedAt:      fixedTime,
	}

	tests := []struct {
		name               string
		req                *proto.Auth
		data               [][]byte
		timeWindow         int64
		findUserStub       *stubFindUserByIDFunc
		expectedUser       *service.User
		expectedErr        error
		checkErrorContains bool
	}{
		{
			name: "Successful authentication",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: timestamp,
				Signature: signature,
				Id:        secretID,
			},
			data:         data,
			timeWindow:   60,
			findUserStub: &stubFindUserByIDFunc{User: user, Err: nil},
			expectedUser: user,
			expectedErr:  nil,
		},
		{
			name: "Empty signature",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: timestamp,
				Signature: []byte{},
				Id:        secretID,
			},
			data:         data,
			timeWindow:   60,
			findUserStub: &stubFindUserByIDFunc{},
			expectedUser: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "Signature is required"),
		},
		{
			name: "Timestamp too old",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: timestamp - 61,
				Signature: signature,
				Id:        secretID,
			},
			data:         data,
			timeWindow:   60,
			findUserStub: &stubFindUserByIDFunc{},
			expectedUser: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "Timestamp is out of range"),
		},
		{
			name: "Timestamp too new",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: timestamp + 61,
				Signature: signature,
				Id:        secretID,
			},
			data:         data,
			timeWindow:   60,
			findUserStub: &stubFindUserByIDFunc{},
			expectedUser: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "Timestamp is out of range"),
		},
		{
			name: "User not found",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: timestamp,
				Signature: signature,
				Id:        secretID,
			},
			data:         data,
			timeWindow:   60,
			findUserStub: &stubFindUserByIDFunc{User: nil, Err: service.ErrNotFound},
			expectedUser: nil,
			expectedErr:  status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "Invalid public key",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: timestamp,
				Signature: signature,
				Id:        secretID,
			},
			data:       data,
			timeWindow: 60,
			findUserStub: &stubFindUserByIDFunc{
				User: &service.User{
					ID:             userID,
					Fingerprint:    fingerprint,
					PublicKey:      "invalid-public-key",
					PublicKeyBytes: []byte("invalid"),
					CreatedAt:      fixedTime,
					UpdatedAt:      fixedTime,
				},
				Err: nil,
			},
			expectedUser:       nil,
			expectedErr:        status.Error(codes.InvalidArgument, "failed to parse public key"),
			checkErrorContains: true,
		},
		{
			name: "Invalid signature",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: timestamp,
				Signature: []byte("invalid-signature"),
				Id:        secretID,
			},
			data:         data,
			timeWindow:   60,
			findUserStub: &stubFindUserByIDFunc{User: user, Err: nil},
			expectedUser: nil,
			expectedErr:  status.Error(codes.Unauthenticated, "invalid signature"),
		},
		{
			name: "Database error",
			req: &proto.Auth{
				UserId:    userID,
				Timestamp: timestamp,
				Signature: signature,
				Id:        secretID,
			},
			data:         data,
			timeWindow:   60,
			findUserStub: &stubFindUserByIDFunc{User: nil, Err: errDB},
			expectedUser: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("failed to find user: %w", errDB).Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServiceServer(
				WithFindUserFunc(tt.findUserStub.FindUserByID),
				WithTimeWindow(tt.timeWindow),
			)
			user, err := s.findAndAuthUser(t.Context(), tt.req, tt.data...)
			if tt.expectedErr == nil {
				require.NoError(t, err, "expected no error")
			} else {
				assert.Error(t, err, "expected an error")
				if tt.checkErrorContains {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				} else {
					assert.Equal(t, tt.expectedErr, err, "expected error %v, got %v", tt.expectedErr, err)
				}
			}
			assert.Equal(t, tt.expectedUser, user, "expected user %v, got %v", tt.expectedUser, user)
		})
	}
}
