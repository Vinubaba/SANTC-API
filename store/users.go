package store

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	UserId   string
	Email    string
	Password string
}

func (s *Store) AddUser(tx *gorm.DB, user User) (id string, err error) {
	db := s.dbOrTx(tx)

	user.UserId = s.StringGenerator.GenerateUuid()

	if err := db.Exec("INSERT INTO users (user_id, email, password) VALUES (?, ?, crypt(?, gen_salt('bf',8)))", user.UserId, user.Email, user.Password).Error; err != nil {
		return "", err
	}

	return user.UserId, nil
}

func (s *Store) userExists(tx *gorm.DB, userID string) bool {
	db := s.dbOrTx(tx)

	u := User{UserId: userID}
	return !db.Model(User{}).Where("user_id = ?", userID).First(&u).RecordNotFound()
}

func (s *Store) CheckUserCredentials(tx *gorm.DB, user User) (User, error) {
	db := s.dbOrTx(tx)

	res := db.Where("email = ? AND password = crypt(?, password)", user.Email, user.Password).First(&user)
	if res.RecordNotFound() {
		return User{}, ErrUserNotFound
	}
	if err := res.Error; err != nil {
		return User{}, err
	}

	return user, nil
}
