package store

import (
	"errors"
	"fmt"
	"strings"

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

type Roles []Role

func (r *Roles) Scan(src interface{}) error {
	switch v := src.(type) {
	case string:
		roles := strings.Split(v, ",")
		for _, role := range roles {
			*r = append(*r, Role{Role: role})
		}
	default:
		return errors.New("need string with roles separated by virgula")
	}
	return nil
}

func (r Roles) ToList() []string {
	roles := make([]string, 0)
	for _, role := range r {
		roles = append(roles, role.Role)
	}
	return roles
}

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

type PendingConnexionRole struct {
	Email string
	Role  string
}

func (s *Store) GetPendingConnexionRoles(tx *gorm.DB, email string) ([]PendingConnexionRole, error) {
	db := s.dbOrTx(tx)

	proles := []PendingConnexionRole{}
	if err := db.Model(PendingConnexionRole{}).Where("email = ?", email).Find(&proles).Error; err != nil {
		return nil, err
	}
	return proles, nil
}

func (s *Store) CreatePendingConnexionRole(tx *gorm.DB, role PendingConnexionRole) error {
	db := s.dbOrTx(tx)
	fmt.Println("create pending connexion role...")
	return db.Model(PendingConnexionRole{}).Create(&role).Error
}

func (s *Store) DeletePendingConnexionRole(tx *gorm.DB, role PendingConnexionRole) error {
	db := s.dbOrTx(tx)
	return db.Where("email = ?", role.Email).Delete(PendingConnexionRole{}).Error
}
