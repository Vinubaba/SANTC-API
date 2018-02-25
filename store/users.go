package store

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	UserId    sql.NullString
	Email     sql.NullString
	FirstName sql.NullString
	LastName  sql.NullString
	Phone     sql.NullString
	Address_1 sql.NullString
	Address_2 sql.NullString
	City      sql.NullString
	State     sql.NullString
	Zip       sql.NullString
	Gender    sql.NullString
	ImageUri  sql.NullString
	Roles     Roles `sql:"-"`
}

func (u *User) Is(role string) bool {
	for _, r := range u.Roles {
		if r.Role == role {
			return true
		}
	}
	return false
}

func (s *Store) AddUser(tx *gorm.DB, user User) (User, error) {
	db := s.dbOrTx(tx)

	user.UserId = sql.NullString{String: s.StringGenerator.GenerateUuid(), Valid: true}
	if err := db.Create(&user).Error; err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *Store) GetUser(tx *gorm.DB, userId string) (User, error) {
	db := s.dbOrTx(tx)
	rows, err := db.Table("users").
		Select("users.user_id, "+
			"users.email, "+
			"users.first_name,"+
			"users.last_name,"+
			"users.phone,"+
			"users.address_1,"+
			"users.address_2,"+
			"users.city,"+
			"users.state,"+
			"users.zip,"+
			"users.gender,"+
			"users.image_uri,"+
			"string_agg(roles.role, ',')").
		Joins("left join roles ON roles.user_id = users.user_id").
		Where("users.user_id = ?", userId).
		Group("users.user_id").
		Rows()
	if err != nil {
		return User{}, err
	}
	users, err := s.scanUserRows(rows)
	if err != nil {
		return User{}, err
	}

	if len(users) > 0 {
		return users[0], nil
	}
	return User{}, ErrUserNotFound
}

func (s *Store) GetUserByEmail(tx *gorm.DB, email string) (User, error) {
	db := s.dbOrTx(tx)
	rows, err := db.Table("users").
		Select("users.user_id,"+
			"users.email,"+
			"users.first_name,"+
			"users.last_name,"+
			"users.phone,"+
			"users.address_1,"+
			"users.address_2,"+
			"users.city,"+
			"users.state,"+
			"users.zip,"+
			"users.gender,"+
			"users.image_uri,"+
			"string_agg(roles.role, ',')").
		Joins("left join roles ON roles.user_id = users.user_id").
		Where("users.email = ?", email).
		Group("users.user_id").
		Rows()
	if err != nil {
		return User{}, err
	}
	users, err := s.scanUserRows(rows)
	if err != nil {
		return User{}, err
	}

	if len(users) > 0 {
		return users[0], nil
	}
	return User{}, ErrUserNotFound
}

func (s *Store) UpdateUser(tx *gorm.DB, user User) (User, error) {
	db := s.dbOrTx(tx)

	res := db.Where("user_id = ?", user.UserId).Model(&User{}).Updates(&user).First(&user)
	if err := res.Error; err != nil {
		return User{}, err
	}
	if res.RecordNotFound() {
		return User{}, ErrUserNotFound
	}

	if err := db.Where("user_id = ?", user.UserId).Model(&Role{}).Find(&user.Roles).Error; err != nil {
		return User{}, errors.Wrap(err, "failed to get roles")
	}

	return user, nil
}

func (s *Store) DeleteUser(tx *gorm.DB, userId string) (err error) {
	db := s.dbOrTx(tx)

	if _, err := s.GetUser(db, userId); err != nil {
		return err
	}

	if err := db.Where("user_id = ?", userId).Delete(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) ListUsers(tx *gorm.DB, roleConstraint string) ([]User, error) {
	db := s.dbOrTx(tx)

	query := db.Table("users").
		Select("users.user_id, " +
			"users.email, " +
			"users.first_name," +
			"users.last_name," +
			"users.phone," +
			"users.address_1," +
			"users.address_2," +
			"users.city," +
			"users.state," +
			"users.zip," +
			"users.gender," +
			"users.image_uri," +
			"string_agg(roles.role, ',')").
		Joins("left join roles ON roles.user_id = users.user_id").
		Group("users.user_id")

	if roleConstraint != "" {
		query = query.Having("string_agg(roles.role, ',') LIKE '%" + roleConstraint + "%'")
	}

	rows, err := query.Rows()
	if err != nil {
		return []User{}, err
	}
	return s.scanUserRows(rows)
}

func (s *Store) scanUserRows(rows *sql.Rows) ([]User, error) {
	users := []User{}
	for rows.Next() {
		currentUser := User{}
		if err := rows.Scan(&currentUser.UserId,
			&currentUser.Email,
			&currentUser.FirstName,
			&currentUser.LastName,
			&currentUser.Phone,
			&currentUser.Address_1,
			&currentUser.Address_2,
			&currentUser.City,
			&currentUser.State,
			&currentUser.Zip,
			&currentUser.Gender,
			&currentUser.ImageUri,
			&currentUser.Roles); err != nil {
			return []User{}, err
		}
		for i := range currentUser.Roles {
			currentUser.Roles[i].UserId = currentUser.UserId.String
		}
		users = append(users, currentUser)
	}

	return users, nil
}
