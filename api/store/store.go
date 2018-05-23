package store

import (
	"database/sql"
	"firebase.google.com/go/auth"
	"github.com/Vinubaba/SANTC-API/api/shared"
	"github.com/jinzhu/gorm"
)

type Store struct {
	Db              *gorm.DB `inject:""`
	StringGenerator interface {
		GenerateUuid() string
	} `inject:""`
	FirebaseClient *auth.Client      `inject:""`
	Config         *shared.AppConfig `inject:""`
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

type SearchOptions struct {
	DaycareId     string
	TeacherId     string
	ResponsibleId string
}
