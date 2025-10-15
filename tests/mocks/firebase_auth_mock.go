package mocks

import (
	"context"

	"firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/mock"
)

// MockFirebaseAuthService es un mock del Firebase Auth Client para testing
type MockFirebaseAuthService struct {
	mock.Mock
}

func (m *MockFirebaseAuthService) CreateUser(ctx context.Context, params *auth.UserToCreate) (*auth.UserRecord, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserRecord), args.Error(1)
}

func (m *MockFirebaseAuthService) CreateCustomToken(ctx context.Context, uid string, claims map[string]interface{}) (string, error) {
	args := m.Called(ctx, uid, claims)
	return args.String(0), args.Error(1)
}

func (m *MockFirebaseAuthService) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Token), args.Error(1)
}

func (m *MockFirebaseAuthService) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserRecord), args.Error(1)
}

func (m *MockFirebaseAuthService) UpdateUser(ctx context.Context, uid string, user *auth.UserToUpdate) (*auth.UserRecord, error) {
	args := m.Called(ctx, uid, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserRecord), args.Error(1)
}

func (m *MockFirebaseAuthService) DeleteUser(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

func (m *MockFirebaseAuthService) SetCustomUserClaims(ctx context.Context, uid string, customClaims map[string]interface{}) error {
	args := m.Called(ctx, uid, customClaims)
	return args.Error(0)
}
