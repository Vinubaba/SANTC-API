package mocks

import (
	"arthurgustin.fr/teddycare/store"
	"context"
	"github.com/stretchr/testify/mock"
)

// Mock that contains also the real implementation, because a real db is used for test
// We only want to mock certain behavior
type MockStore struct {
	mock.Mock
	Store *store.Store
}

func (s *MockStore) BeginTransaction() {
	s.Store.BeginTransaction()
}

func (s *MockStore) Rollback() {
	s.Store.Rollback()
}

func (s *MockStore) Commit() {
	s.Store.Commit()
}

func (s *MockStore) AddUser(ctx context.Context) (id string, err error) {
	args := s.Called()
	return args.Get(0).(string), args.Error(1)
}

func (s *MockStore) AddAdultResponsible(ctx context.Context, adult store.AdultResponsible) (id string, err error) {
	args := s.Called()
	return args.Get(0).(string), args.Error(1)
}

func (s *MockStore) AddChild(ctx context.Context, child store.Child) (id string, err error) {
	args := s.Called()
	return args.Get(0).(string), args.Error(1)
}

func (s *MockStore) SetResponsible(ctx context.Context, responsibleOf store.ResponsibleOf) error {
	args := s.Called()
	return args.Error(0)
}

func (s *MockStore) ListAdultResponsible(ctx context.Context) ([]store.AdultResponsible, error) {
	args := s.Called()
	return args.Get(0).([]store.AdultResponsible), args.Error(1)
}
