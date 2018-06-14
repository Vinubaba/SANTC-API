package store

import (
	"database/sql"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/Vinubaba/SANTC-API/api/shared"
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
	DaycareId sql.NullString
}

type TeacherClass struct {
	TeacherId sql.NullString
	ClassId   sql.NullString
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

func (s *Store) GetUser(tx *gorm.DB, userId string, searchOptions SearchOptions) (User, error) {
	db := s.dbOrTx(tx)

	query := db.Table("users").
		Select("users.user_id, " +
			"users.daycare_id," +
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
		Joins("left join roles ON roles.user_id = users.user_id")

	if searchOptions.DaycareId != "" {
		query = query.Where("users.daycare_id = ?", searchOptions.DaycareId)
	}
	query = query.Where("users.user_id = ?", userId).Group("users.user_id")

	rows, err := query.Rows()
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
		Select("users.user_id, "+
			"users.daycare_id,"+
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

	if _, err := s.GetUser(db, userId, SearchOptions{}); err != nil {
		return err
	}

	if err := db.Where("user_id = ?", userId).Delete(&User{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *Store) ListDaycareUsers(tx *gorm.DB, roleConstraint string, options SearchOptions) ([]User, error) {
	db := s.dbOrTx(tx)
	query := db.Table("users, children, teacher_classes, roles").
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
			"string_agg(roles.role, ',')," +
			"users.daycare_id")
	if options.DaycareId != "" {
		query = query.Where("users.daycare_id = ?", options.DaycareId)
	}
	if len(options.ChildrenId) > 0 {
		if roleConstraint == shared.ROLE_TEACHER {
			for _, childId := range options.ChildrenId {
				query = query.Where("children.child_id = ?", childId)
			}
			query = query.Where("children.class_id = teacher_classes.class_id")
			query = query.Where("users.user_id = teacher_classes.teacher_id")
		}
	}
	query = query.Where("roles.user_id = users.user_id").Group("users.user_id")
	/*query = query.Joins("left join roles ON roles.user_id = users.user_id").
		Group("users.user_id")*/

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
			&currentUser.DaycareId,
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

func (s *Store) SetTeacherClass(tx *gorm.DB, teacherClass TeacherClass) error {
	db := s.dbOrTx(tx)

	if err := db.Create(teacherClass).Error; err != nil {
		return err
	}
	return nil
}
