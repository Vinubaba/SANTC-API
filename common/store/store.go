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

func DbNullString(value *string) sql.NullString {
	// will update value in db
	if value != nil {
		return sql.NullString{
			String: *value,
			Valid:  true,
		}
	}
	// will ignore this value
	return sql.NullString{
		Valid: false,
	}
}

func DbNullBool(value *bool) sql.NullBool {
	// will update value in db
	if value != nil {
		return sql.NullBool{
			Bool:  *value,
			Valid: true,
		}
	}
	// will ignore this value
	return sql.NullBool{
		Valid: false,
	}
}

func DbNullInt64(value *int64) sql.NullInt64 {
	// will update value in db
	if value != nil {
		return sql.NullInt64{
			Int64: *value,
			Valid: true,
		}
	}
	// will ignore this value
	return sql.NullInt64{
		Valid: false,
	}
}

func (s *Store) newId() sql.NullString {
	id := s.StringGenerator.GenerateUuid()
	return DbNullString(&id)
}

type SearchOptions struct {
	DaycareId     string
	TeacherId     string
	ResponsibleId string
	ChildrenId    []string
}
