package store

import (
	"firebase.google.com/go/auth"
	"github.com/jinzhu/gorm"
)

type Store struct {
	Db              *gorm.DB `inject:""`
	StringGenerator interface {
		GenerateUuid() string
	} `inject:""`
	FirebaseClient *auth.Client `inject:""`
}

func (s *Store) Tx() *gorm.DB {
	return s.Db.Begin()
}

func (s *Store) dbOrTx(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return s.Db
}
