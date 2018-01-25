package store

import "github.com/jinzhu/gorm"

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
