package store

import (
	"database/sql"
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

func DbNullString(value string) sql.NullString {
	if value != "" {
		return sql.NullString{
			String: value,
			Valid:  true,
		}
	}
	return sql.NullString{
		String: "",
		Valid:  false,
	}
}
