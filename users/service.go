package users

import (
	"context"
	"strings"

	"github.com/Vinubaba/SANTC-API/shared"
	"github.com/Vinubaba/SANTC-API/storage"
	"github.com/Vinubaba/SANTC-API/store"

	"github.com/Vinubaba/SANTC-API/claims"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrInvalidEmail           = errors.New("invalid email")
	ErrInvalidPasswordFormat  = errors.New("password must be at least 6 characters long")
	ErrCreateDifferentDaycare = errors.New("cannot create user for another daycare")
	ErrUserNotOfficeManager   = errors.New("you must be an office manager to perform this operation")
)

type Service interface {
	AddUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error)
	GetUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error)
	UpdateUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error)
	DeleteUserByRoles(ctx context.Context, request UserTransport, roles ...string) error
	ListUsersByRole(ctx context.Context, roleConstraint string) ([]store.User, error)

	SetTeacherClass(ctx context.Context, teacherId, classId string) error
}

type UserService struct {
	Store interface {
		// User methods
		AddUser(tx *gorm.DB, user store.User) (store.User, error)
		ListDaycareUsers(tx *gorm.DB, roleConstraint string, daycareId string) ([]store.User, error)
		UpdateUser(tx *gorm.DB, user store.User) (store.User, error)
		DeleteUser(tx *gorm.DB, userId string) (err error)
		GetUser(tx *gorm.DB, userId string, searchOptions store.SearchOptions) (store.User, error)
		GetUserByEmail(tx *gorm.DB, email string) (store.User, error)

		// Teacher specific method
		SetTeacherClass(tx *gorm.DB, teacherClass store.TeacherClass) error
		GetClass(tx *gorm.DB, classId string, options store.SearchOptions) (store.Class, error)

		AddRole(tx *gorm.DB, role store.Role) (store.Role, error)
		Tx() *gorm.DB
	} `inject:""`
	FirebaseClient interface {
		DeleteUserByEmail(ctx context.Context, email string) error
	} `inject:"teddyFirebaseClient"`
	Storage storage.Storage   `inject:""`
	Config  *shared.AppConfig `inject:""`
	Logger  *shared.Logger    `inject:""`
}

func (c *UserService) validateDaycareRequest(ctx context.Context, request *UserTransport) error {
	claims := ctx.Value("claims").(map[string]interface{})

	if claims[shared.ROLE_ADMIN].(bool) && request.DaycareId == "" {
		return errors.New("as an admin, you must specify the user daycare")
	}

	// default to requester daycare (e.g office manager)
	if request.DaycareId == "" {
		request.DaycareId = claims["daycareId"].(string)
	}

	if claims["daycareId"].(string) != request.DaycareId {
		return ErrCreateDifferentDaycare
	}

	return nil
}

func (c *UserService) AddUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error) {
	var err error
	if err = c.validateDaycareRequest(ctx, &request); err != nil {
		return store.User{}, ErrCreateDifferentDaycare
	}

	tx := c.Store.Tx()
	if tx.Error != nil {
		return store.User{}, errors.Wrap(tx.Error, "failed to create user")
	}

	request.ImageUri, err = c.Storage.Store(ctx, request.ImageUri)
	if err != nil {
		return store.User{}, err
	}

	if request.DaycareId == "" {
		request.DaycareId = c.Config.PublicDaycareId
	}

	createdUser, err := c.Store.AddUser(tx, store.User{
		Email:     store.DbNullString(request.Email),
		FirstName: store.DbNullString(request.FirstName),
		LastName:  store.DbNullString(request.LastName),
		Gender:    store.DbNullString(request.Gender),
		Zip:       store.DbNullString(request.Zip),
		State:     store.DbNullString(request.State),
		Phone:     store.DbNullString(request.Phone),
		City:      store.DbNullString(request.City),
		Address_1: store.DbNullString(request.Address_1),
		Address_2: store.DbNullString(request.Address_2),
		UserId:    store.DbNullString(request.Id),
		ImageUri:  store.DbNullString(request.ImageUri),
		DaycareId: store.DbNullString(request.DaycareId),
	})
	if err != nil {
		tx.Rollback()
		return store.User{}, errors.Wrap(err, "failed to create user")
	}

	for _, role := range roles {
		_, err := c.Store.AddRole(tx, store.Role{
			Role:   role,
			UserId: createdUser.UserId.String,
		})
		if err != nil {
			tx.Rollback()
			return store.User{}, errors.Wrap(err, "failed to set user role")
		}
		createdUser.Roles = append(createdUser.Roles, store.Role{
			UserId: createdUser.UserId.String,
			Role:   role,
		})
	}

	tx.Commit()
	c.setBucketUri(ctx, &createdUser)
	return createdUser, nil
}

func (c *UserService) UpdateUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error) {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	user, err := c.Store.GetUser(nil, request.Id, searchOptions)
	if err != nil {
		return store.User{}, errors.Wrap(err, "failed to update user")
	}

	for _, role := range roles {
		if !user.Is(role) {
			return store.User{}, errors.Errorf("user %s it not a %s", user.UserId, role)
		}
	}

	request.ImageUri, err = c.Storage.Store(ctx, request.ImageUri)
	if err != nil {
		return store.User{}, err
	}
	user, err = c.Store.UpdateUser(nil, store.User{
		UserId:    store.DbNullString(request.Id),
		Email:     store.DbNullString(request.Email),
		Address_1: store.DbNullString(request.Address_1),
		Address_2: store.DbNullString(request.Address_2),
		City:      store.DbNullString(request.City),
		State:     store.DbNullString(request.State),
		Zip:       store.DbNullString(request.Zip),
		Phone:     store.DbNullString(request.Phone),
		Gender:    store.DbNullString(request.Gender),
		LastName:  store.DbNullString(request.LastName),
		FirstName: store.DbNullString(request.FirstName),
		ImageUri:  store.DbNullString(request.ImageUri),
		DaycareId: store.DbNullString(request.DaycareId),
	})
	if err != nil {
		return store.User{}, err
	}

	c.setBucketUri(ctx, &user)
	return user, nil
}

func (c *UserService) setBucketUri(ctx context.Context, user *store.User) {
	if user.ImageUri.String == "" {
		return
	}
	if !strings.Contains(user.ImageUri.String, "/") {
		uri, err := c.Storage.Get(ctx, user.ImageUri.String)
		if err != nil {
			c.Logger.Warn(ctx, "failed to generate image uri", "err", err.Error())
			user.ImageUri.Scan("")
		}
		user.ImageUri.Scan(uri)
	}
}

func (c *UserService) GetUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error) {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	user, err := c.Store.GetUser(nil, request.Id, searchOptions)
	if err != nil {
		return store.User{}, errors.Wrap(err, "failed to get user")
	}

	for _, role := range roles {
		if !user.Is(role) {
			return store.User{}, errors.Errorf("user %s is not a %s", user.UserId.String, role)
		}
	}

	c.setBucketUri(ctx, &user)

	return user, nil
}

// Only called by firebase middleware
func (c *UserService) GetUserByEmail(ctx context.Context, request UserTransport) (store.User, error) {
	user, err := c.Store.GetUserByEmail(nil, request.Email)
	if err != nil {
		return store.User{}, errors.Wrap(err, "failed to get user")
	}

	c.setBucketUri(ctx, &user)

	return user, nil
}

func (c *UserService) DeleteUserByRoles(ctx context.Context, request UserTransport, roles ...string) error {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	user, err := c.Store.GetUser(nil, request.Id, searchOptions)
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	for _, role := range roles {
		if !user.Is(role) {
			return errors.Errorf("user %s it not %s", user.UserId, role)
		}
	}

	if err := c.FirebaseClient.DeleteUserByEmail(ctx, user.Email.String); err != nil {
		c.Logger.Warn(ctx, "failed to delete user from firebase")
	}

	if err := c.Store.DeleteUser(nil, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if err := c.Storage.Delete(ctx, user.ImageUri.String); err != nil {
		c.Logger.Warn(ctx, "failed to delete user image", "imageUri", user.ImageUri.String, "err", err.Error())
	}

	return nil
}

func (c *UserService) ListUsersByRole(ctx context.Context, roleConstraint string) ([]store.User, error) {
	claims := ctx.Value("claims").(map[string]interface{})
	daycareId := ""
	if !claims[shared.ROLE_ADMIN].(bool) {
		daycareId = claims["daycareId"].(string)
	}

	users, err := c.Store.ListDaycareUsers(nil, roleConstraint, daycareId)
	if err != nil {
		return make([]store.User, 0), err
	}

	for i := range users {
		c.setBucketUri(ctx, &users[i])
	}
	return users, nil
}

func (c *UserService) SetTeacherClass(ctx context.Context, teacherId, classId string) error {
	searchOptions := claims.GetDefaultSearchOptions(ctx)
	if _, err := c.Store.GetClass(nil, classId, searchOptions); err != nil {
		return err
	}
	if _, err := c.Store.GetUser(nil, teacherId, searchOptions); err != nil {
		return err
	}

	if err := c.Store.SetTeacherClass(nil, store.TeacherClass{
		TeacherId: store.DbNullString(teacherId),
		ClassId:   store.DbNullString(classId),
	}); err != nil {
		return err
	}

	return nil
}

// ServiceMiddleware is a chainable behavior modifier for adultResponsibleService.
type ServiceMiddleware func(UserService) UserService
