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
	args := m.Called()
	return args.Error(0)
}
