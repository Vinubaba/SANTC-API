package store

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

const (
	ROLE_ADMIN          = "admin"
	ROLE_TEACHER        = "teacher"
	ROLE_ADULT          = "adult"
	ROLE_OFFICE_MANAGER = "officemanager"
)

var (
	roles = []string{ROLE_ADMIN, ROLE_TEACHER, ROLE_ADULT, ROLE_OFFICE_MANAGER}
)

type Role struct {
	UserId string
	Role   string
}

func (s *Store) AddRole(tx *gorm.DB, role Role) (Role, error) {
	db := s.dbOrTx(tx)

	if !s.isRoleValid(role.Role) {
		return Role{}, fmt.Errorf("role is not valid, must be %s", roles)
	}

	if err := db.Create(&role).Error; err != nil {
		return Role{}, err
	}
	return role, nil
}

func (s *Store) isRoleValid(role string) bool {
	for _, r := range roles {
		if role == r {
			return true
		}
	}
	return false
}

func (s *Store) GetUserRoles(tx *gorm.DB, userId string) ([]Role, error) {
	db := s.dbOrTx(tx)

	var roles []Role
	if err := db.Where("user_id = ?", userId).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
