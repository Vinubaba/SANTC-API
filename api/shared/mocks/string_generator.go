package shared

import (
	"github.com/stretchr/testify/mock"
)

type MockStringGenerator struct {
	mock.Mock
}

func (m *MockStringGenerator) GenerateRandomName() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockStringGenerator) GenerateUuid() string {
	args := m.Called()
	return args.Get(0).(string)
}
