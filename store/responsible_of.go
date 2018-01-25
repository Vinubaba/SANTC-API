package store

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
)

var (
	allRelationships       = []string{REL_FATHER, REL_MOTHER, REL_GRANDFATHER, REL_GRANDMOTHER, REL_GUARDIAN}
	ErrInvalidRelationship = errors.New(fmt.Sprintf("relationship is not valid, it should be one of %s", allRelationships))
)

const (
	REL_MOTHER      = "mother"
	REL_FATHER      = "father"
	REL_GRANDMOTHER = "grandmother"
	REL_GRANDFATHER = "grandfather"
	REL_GUARDIAN    = "guardian"
)

type ResponsibleOf struct {
	ResponsibleId string `sql:"type:varchar(64) REFERENCES users(user_id)"`
	ChildId       string `sql:"type:varchar(64) REFERENCES users(user_id)"`
	Relationship  string
}

func (ResponsibleOf) TableName() string {
	return "responsible_of"
}

func (s *Store) SetResponsible(ctx context.Context, responsibleOf ResponsibleOf) error {
	if !s.isRelationshipValid(responsibleOf.Relationship) {
		return errors.Wrap(ErrInvalidRelationship, fmt.Sprintf("relationship %s is not valid", responsibleOf.Relationship))
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
