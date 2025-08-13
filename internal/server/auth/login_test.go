package auth

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"gophkeeper/internal/server/service"
	"gophkeeper/pkg/crypt"
	"gophkeeper/pkg/proto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stubFindUserByFingerprintFunc creates a stub for FindUserByFingerprintFunc with configurable return values.
type stubFindUserByFingerprintFunc struct {
	User *service.User
	Err  error
}

func (s *stubFindUserByFingerprintFunc) FindUserByFingerprint(context.Context, string) (*service.User, error) {
	return s.User, s.Err
}

func TestServiceServer_Login(t *testing.T) {
	fixedTime := time.Now().Truncate(time.Second).UTC()

	privateKey, publicKey := crypt.GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	assert.NoError(t, err)
	fingerprint := crypt.Fingerprint(publicKeyBytes)
	encodedPublicKey := crypt.EncodePublicKey(publicKeyBytes)
	timestamp := fixedTime.Unix()
	signature, err := crypt.SignPSSWithTimestamp(privateKey, timestamp, publicKeyBytes)
	assert.NoError(t, err)
	user := &service.User{
		ID:             1,
		Fingerprint:    fingerprint,
		PublicKey:      encodedPublicKey,
		PublicKeyBytes: publicKeyBytes,
		CreatedAt:      fixedTime,
		UpdatedAt:      fixedTime,
	}
	errDB := errors.New("database error")

	tests := []struct {
		name               string
		req                *proto.LoginRequest
		findStub           *stubFindUserByFingerprintFunc
		expectedResponse   *proto.AuthResponse
		expectedErr        error
		checkErrorContains bool
	}{
		{
			name: "Successful login",
			req: &proto.LoginRequest{
				Fingerprint: fingerprint,
				Timestamp:   timestamp,
				Signature:   signature,
			},
			findStub: &stubFindUserByFingerprintFunc{
				User: user,
				Err:  nil,
			},
			expectedResponse: &proto.AuthResponse{
				UserId: 1,
			},
			expectedErr: nil,
		},
		{
			name: "Empty fingerprint",
			req: &proto.LoginRequest{
				Fingerprint: "",
				Timestamp:   timestamp,
				Signature:   signature,
			},
			findStub:         &stubFindUserByFingerprintFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "Fingerprint is required"),
		},
		{
			name: "Empty signature",
			req: &proto.LoginRequest{
				Fingerprint: fingerprint,
				Timestamp:   timestamp,
				Signature:   []byte{},
			},
			findStub:         &stubFindUserByFingerprintFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "Signature is required"),
		},
		{
			name: "Timestamp out of range (too old)",
			req: &proto.LoginRequest{
				Fingerprint: fingerprint,
				Timestamp:   timestamp - defaultTimeWindow - 10,
				Signature:   signature,
			},
			findStub: &stubFindUserByFingerprintFunc{
				User: user,
				Err:  nil,
			},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "Timestamp is out of range"),
		},
		{
			name: "Timestamp out of range (too new)",
			req: &proto.LoginRequest{
				Fingerprint: fingerprint,
				Timestamp:   timestamp + defaultTimeWindow + 10,
				Signature:   signature,
			},
			findStub: &stubFindUserByFingerprintFunc{
				User: user,
				Err:  nil,
			},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "Timestamp is out of range"),
		},
		{
			name: "User not found",
			req: &proto.LoginRequest{
				Fingerprint: fingerprint,
				Timestamp:   timestamp,
				Signature:   signature,
			},
			findStub: &stubFindUserByFingerprintFunc{
				User: nil,
				Err:  service.ErrNotFound,
			},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.NotFound, "user not found"),
		},
		{
			name: "Invalid public key",
			req: &proto.LoginRequest{
				Fingerprint: fingerprint,
				Timestamp:   timestamp,
				Signature:   signature,
			},
			findStub: &stubFindUserByFingerprintFunc{
				User: &service.User{
					ID:             1,
					Fingerprint:    fingerprint,
					PublicKey:      "invalid-public-key",
					PublicKeyBytes: []byte("invalid"),
					CreatedAt:      fixedTime,
					UpdatedAt:      fixedTime,
				},
				Err: nil,
			},
			expectedResponse:   nil,
			expectedErr:        status.Error(codes.InvalidArgument, "failed to parse public key"),
			checkErrorContains: true,
		},
		{
			name: "Invalid signature",
			req: &proto.LoginRequest{
				Fingerprint: fingerprint,
				Timestamp:   timestamp,
				Signature:   []byte("invalid-signature"),
			},
			findStub: &stubFindUserByFingerprintFunc{
				User: user,
				Err:  nil,
			},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "invalid signature"),
		},
		{
			name: "Database error",
			req: &proto.LoginRequest{
				Fingerprint: fingerprint,
				Timestamp:   timestamp,
				Signature:   signature,
			},
			findStub: &stubFindUserByFingerprintFunc{
				User: nil,
				Err:  errDB,
			},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.Internal, fmt.Errorf("failed to get user by fingerprint: %w", errDB).Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServiceServer(WithFindUserFunc(tt.findStub.FindUserByFingerprint))
			resp, err := s.Login(t.Context(), tt.req)
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
			if tt.expectedResponse != nil {
				assert.Equal(t, tt.expectedResponse.UserId, resp.UserId, "expected response %v, got %v", tt.expectedResponse, resp)
			}
		})
	}
}
