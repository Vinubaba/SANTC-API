package mocks

import (
	"context"

	"firebase.google.com/go/auth"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) DeleteUserByEmail(ctx context.Context, email string) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClient) VerifyIDToken(idToken string) (*auth.Token, error) {
	args := m.Called()
	return args.Get(0).(*auth.Token), args.Error(1)
}

func (m *MockClient) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	args := m.Called()
	return args.Get(0).(*auth.UserRecord), args.Error(1)
}

func (m *MockClient) SetCustomUserClaims(ctx context.Context, uid string, customClaims map[string]interface{}) error {
	args := m.Called()
	return args.Error(0)
}
