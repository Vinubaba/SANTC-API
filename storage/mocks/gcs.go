package mocks

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockGcs struct {
	mock.Mock
}

func (m *MockGcs) Store(ctx context.Context, encodedImage, mimeType string) (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}

func (m *MockGcs) Get(ctx context.Context, filename string) (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}

func (m *MockGcs) Delete(ctx context.Context, filename string) error {
	args := m.Called()
	return args.Error(0)
}
