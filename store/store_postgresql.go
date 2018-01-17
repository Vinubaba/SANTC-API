package store

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

type Store struct {
	Db              *gorm.DB `inject:""`
	transaction     *gorm.DB
	StringGenerator interface {
		GenerateUuid() string
	} `inject:""`
}

func (s *Store) BeginTransaction() {
	s.transaction = s.Db.Begin()
}

func (s *Store) Rollback() {
	if s.transaction != nil {
		s.transaction.Rollback()
	}
}

func (s *Store) Commit() {
	if s.transaction != nil {
		s.transaction.Commit()
	}
}

func (s *Store) AddUser(ctx context.Context) (id string, err error) {
	u := User{UserId: uuid.NewV4().String()}
	if err := s.Db.Create(&u).Error; err != nil {
		return "", err
	}
	return u.UserId, nil
}

func (s *Store) AddAdultResponsible(ctx context.Context, adult AdultResponsible) (id string, err error) {
	userId, err := s.AddUser(ctx)
	if err != nil {
		return "", err
	}

	adult.ResponsibleId = userId
	if err := s.Db.Create(&adult).Error; err != nil {
		return "", err
	}

	return adult.ResponsibleId, nil
}

func (s *Store) AddChild(ctx context.Context, child Child) (id string, err error) {
	userId, err := s.AddUser(ctx)
	if err != nil {
		return "", err
	}

	child.ChildId = userId
	if err := s.Db.Create(&child).Error; err != nil {
		return "", err
	}

	return child.ChildId, nil
}

func (s *Store) SetResponsible(ctx context.Context, responsibleOf ResponsibleOf) error {
	if !s.isRelationshipValid(responsibleOf.Relationship) {
		return fmt.Errorf("relationship is not valid, it should be one of %s", allRelationships)
	}

	if err := s.Db.Create(&responsibleOf).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) isRelationshipValid(relationship string) bool {
	for _, rel := range allRelationships {
		if rel == relationship {
			return true
		}
	}
	return false
}

func (s *Store) ListAdultResponsible(ctx context.Context) ([]AdultResponsible, error) {
	adults := []AdultResponsible{}
	if err := s.Db.Find(&adults).Error; err != nil {
		return nil, err
	}

	return adults, nil
}
