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

// stubRegisterUserFunc creates a stub for RegisterUserFunc with configurable return values.
type stubRegisterUserFunc struct {
	UserID int64
	Err    error
}

func (s *stubRegisterUserFunc) RegisterUser(ctx context.Context, publicKeyBytes []byte) (int64, error) {
	return s.UserID, s.Err
}

func TestServiceServer_Register(t *testing.T) {
	fixedTime := time.Now().Truncate(time.Second).UTC()

	privateKey, publicKey := crypt.GenerateTestRSAKeyPair(t)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	assert.NoError(t, err)
	timestamp := fixedTime.Unix()
	signature, err := crypt.SignPSSWithTimestamp(privateKey, timestamp, publicKeyBytes)
	assert.NoError(t, err)
	errDB := errors.New("database error")

	tests := []struct {
		name               string
		req                *proto.RegisterRequest
		registerStub       *stubRegisterUserFunc
		expectedResponse   *proto.AuthResponse
		expectedErr        error
		checkErrorContains bool
	}{
		{
			name: "Successful registration",
			req: &proto.RegisterRequest{
				PublicKeyBytes: publicKeyBytes,
				Timestamp:      timestamp,
				Signature:      signature,
			},
			registerStub: &stubRegisterUserFunc{
				UserID: 1,
				Err:    nil,
			},
			expectedResponse: &proto.AuthResponse{
				UserId: 1,
			},
			expectedErr: nil,
		},
		{
			name: "Empty public key",
			req: &proto.RegisterRequest{
				PublicKeyBytes: []byte{},
				Timestamp:      timestamp,
				Signature:      signature,
			},
			registerStub:     &stubRegisterUserFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "PublicKey is required"),
		},
		{
			name: "Empty signature",
			req: &proto.RegisterRequest{
				PublicKeyBytes: publicKeyBytes,
				Timestamp:      timestamp,
				Signature:      []byte{},
			},
			registerStub:     &stubRegisterUserFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "Signature is required"),
		},
		{
			name: "Timestamp out of range (too old)",
			req: &proto.RegisterRequest{
				PublicKeyBytes: publicKeyBytes,
				Timestamp:      timestamp - defaultTimeWindow - 10,
				Signature:      signature,
			},
			registerStub: &stubRegisterUserFunc{
				UserID: 1,
				Err:    nil,
			},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "Timestamp is out of range"),
		},
		{
			name: "Timestamp out of range (too new)",
			req: &proto.RegisterRequest{
				PublicKeyBytes: publicKeyBytes,
				Timestamp:      timestamp + defaultTimeWindow + 10,
				Signature:      signature,
			},
			registerStub:     &stubRegisterUserFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "Timestamp is out of range"),
		},
		{
			name: "Invalid public key",
			req: &proto.RegisterRequest{
				PublicKeyBytes: []byte("invalid"),
				Timestamp:      timestamp,
				Signature:      signature,
			},
			registerStub:       &stubRegisterUserFunc{},
			expectedResponse:   nil,
			expectedErr:        status.Error(codes.InvalidArgument, "failed to parse public key"),
			checkErrorContains: true,
		},
		{
			name: "Invalid signature",
			req: &proto.RegisterRequest{
				PublicKeyBytes: publicKeyBytes,
				Timestamp:      timestamp,
				Signature:      []byte("invalid-signature"),
			},
			registerStub:     &stubRegisterUserFunc{},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.InvalidArgument, "invalid signature"),
		},
		{
			name: "User already exists",
			req: &proto.RegisterRequest{
				PublicKeyBytes: publicKeyBytes,
				Timestamp:      timestamp,
				Signature:      signature,
			},
			registerStub: &stubRegisterUserFunc{
				UserID: 0,
				Err:    service.ErrConflict,
			},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.AlreadyExists, service.ErrConflict.Error()),
		},
		{
			name: "Database error",
			req: &proto.RegisterRequest{
				PublicKeyBytes: publicKeyBytes,
				Timestamp:      timestamp,
				Signature:      signature,
			},
			registerStub: &stubRegisterUserFunc{
				UserID: 0,
				Err:    errDB,
			},
			expectedResponse: nil,
			expectedErr:      status.Error(codes.Internal, fmt.Errorf("failed to register: %w", errDB).Error()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServiceServer(WithRegisterUserFunc(tt.registerStub.RegisterUser))
			resp, err := s.Register(t.Context(), tt.req)
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
