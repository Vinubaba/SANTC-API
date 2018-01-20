package store

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
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

func (s *Store) AddUser(ctx context.Context, user User) (id string, err error) {
	user.UserId = s.StringGenerator.GenerateUuid()

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("failed to encrypt password")
	}
	user.Password = string(hash)

	if err := s.Db.Create(&user).Error; err != nil {
		return "", err
	}
	return user.UserId, nil
}

func (s *Store) AddAdultResponsible(ctx context.Context, adult AdultResponsible) (id string, err error) {
	if err := s.Db.Create(&adult).Error; err != nil {
		return "", err
	}

	return adult.ResponsibleId, nil
}

func (s *Store) AddChild(ctx context.Context, child Child) (Child, error) {
	child.ChildId = s.StringGenerator.GenerateUuid()

	if err := s.Db.Create(&child).Error; err != nil {
		return Child{}, err
	}

	return child, nil
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

func (s *Store) DeleteAdultResponsible(ctx context.Context, adultId string) (err error) {
	if !s.userExists(ctx, adultId) {
		return ErrUserNotFound
	}

	if err := s.Db.Delete(&User{UserId: adultId}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAdultResponsible(ctx context.Context, adultId string) (AdultResponsible, error) {
	adult := AdultResponsible{}
	res := s.Db.Where("responsible_id = ?", adultId).First(&adult)
	if res.RecordNotFound() {
		return AdultResponsible{}, ErrUserNotFound
	}
	if err := res.Error; err != nil {
		return AdultResponsible{}, err
	}

	return adult, nil
}

func (s *Store) UpdateAdultResponsible(ctx context.Context, adult AdultResponsible) (AdultResponsible, error) {
	userId := adult.ResponsibleId
	email := adult.Email

	user := User{}

	if email != "" {
		res := s.Db.Model(&user).Where("user_id = ?", userId).Update("email", email)
		if res.RecordNotFound() {
			return AdultResponsible{}, ErrUserNotFound
		}
		if err := res.Error; err != nil {
			return AdultResponsible{}, err
		}
	}

	adult.ResponsibleId = ""
	adult.Email = ""

	res := s.Db.Model(&AdultResponsible{}).Where("responsible_id = ?", userId).Updates(adult).First(&adult)
	if err := res.Error; err != nil {
		return AdultResponsible{}, err
	}
	if res.RecordNotFound() {
		return AdultResponsible{}, ErrUserNotFound
	}

	return adult, nil
}

func (s *Store) userExists(ctx context.Context, userID string) bool {
	u := User{UserId: userID}
	return !s.Db.First(&u).RecordNotFound()
}
