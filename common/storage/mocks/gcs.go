package mocks

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

type MockGcs struct {
	mock.Mock
}

func (m *MockGcs) Store(ctx context.Context, b64image string, folder string) (string, error) {
	args := m.Called(ctx, b64image, folder)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockGcs) Get(ctx context.Context, filename string) (string, error) {
	args := m.Called(ctx, filename)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockGcs) Delete(ctx context.Context, filename string) error {
	args := m.Called(ctx, filename)
	return args.Error(0)
}

func (m *MockGcs) CallsForMethod(method string) []mock.Call {
	var calls []mock.Call
	for _, call := range m.Calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

func (m *MockGcs) AssertStoredImage(path string) {
	It("should store the image to the right folder", func() {
		calls := m.CallsForMethod("Store")
		Expect(calls).To(HaveLen(1))
		args := calls[0].Arguments
		Expect(args.String(2)).To(Equal(path))
	})
}

func (m *MockGcs) Reset() {
	m.Mock = mock.Mock{}
}
