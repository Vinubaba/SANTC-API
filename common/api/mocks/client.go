package mocks

import (
	"context"

	"github.com/Vinubaba/SANTC-API/common/api"
	"github.com/stretchr/testify/mock"
)

type MockApiClient struct {
	mock.Mock
}

func (m *MockApiClient) AddImageApprovalRequest(ctx context.Context, approval api.PhotoRequestTransport) error {
	args := m.Called(ctx, approval)
	return args.Error(0)
}
func (m *MockApiClient) GetChild(ctx context.Context, childId string) (api.ChildTransport, error) {
	args := m.Called(ctx, childId)
	return args.Get(0).(api.ChildTransport), args.Error(0)
}
