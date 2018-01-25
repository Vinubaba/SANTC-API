package store

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	UserId   string `gorm:"primary_key:true"`
	Email    string
	Password string
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

func (s *Store) userExists(ctx context.Context, userID string) bool {
	u := User{UserId: userID}
	return !s.Db.Model(User{}).Where("user_id = ?", userID).First(&u).RecordNotFound()
}
