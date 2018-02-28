package users

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Vinubaba/SANTC-API/shared"
	"github.com/Vinubaba/SANTC-API/storage"
	"github.com/Vinubaba/SANTC-API/store"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrInvalidEmail          = errors.New("invalid email")
	ErrInvalidPasswordFormat = errors.New("password must be at least 6 characters long")
)

type Service interface {
	AddUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error)
	GetUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error)
	UpdateUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error)
	DeleteUserByRoles(ctx context.Context, request UserTransport, roles ...string) error
	ListUsersByRole(ctx context.Context, roleConstraint string) ([]store.User, error)
}

type UserService struct {
	Store interface {
		AddUser(tx *gorm.DB, user store.User) (store.User, error)
		ListUsers(tx *gorm.DB, roleConstraint string) ([]store.User, error)
		UpdateUser(tx *gorm.DB, user store.User) (store.User, error)
		DeleteUser(tx *gorm.DB, userId string) (err error)
		GetUser(tx *gorm.DB, userId string) (store.User, error)
		GetUserByEmail(tx *gorm.DB, email string) (store.User, error)

		AddRole(tx *gorm.DB, role store.Role) (store.Role, error)
		Tx() *gorm.DB
	} `inject:""`
	FirebaseClient interface {
		DeleteUserByEmail(ctx context.Context, email string) error
	} `inject:"teddyFirebaseClient"`
	Storage storage.Storage `inject:""`
	Logger  *shared.Logger  `inject:""`
}

func dbString(value string) sql.NullString {
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

func (c *UserService) AddUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error) {
	tx := c.Store.Tx()
	if tx.Error != nil {
		return store.User{}, errors.Wrap(tx.Error, "failed to create user")
	}

	if err := c.setAndStoreDecodedImage(ctx, &request); err != nil {
		return store.User{}, err
	}

	createdUser, err := c.Store.AddUser(tx, store.User{
		Email:     dbString(request.Email),
		FirstName: dbString(request.FirstName),
		LastName:  dbString(request.LastName),
		Gender:    dbString(request.Gender),
		Zip:       dbString(request.Zip),
		State:     dbString(request.State),
		Phone:     dbString(request.Phone),
		City:      dbString(request.City),
		Address_1: dbString(request.Address_1),
		Address_2: dbString(request.Address_2),
		UserId:    dbString(request.Id),
		ImageUri:  dbString(request.ImageUri),
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

func (c *UserService) setAndStoreDecodedImage(ctx context.Context, request *UserTransport) error {
	if strings.HasPrefix(request.ImageUri, "data:image/jpeg;base64,") {
		mimeType := "image/jpeg"
		encoded := strings.TrimPrefix(request.ImageUri, "data:image/jpeg;base64,")

		var err error
		request.ImageUri, err = c.Storage.Store(ctx, encoded, mimeType)
		if err != nil {
			return errors.Wrap(err, "failed to store image")
		}
	}
	return nil
}

func (c *UserService) UpdateUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error) {
	/*
		//todo: when affiliation is developed
		userToUpdate, err := c.Store.GetUser(nil, request.Id)
		if err != nil {
			return store.User{}, err
		}

		userToUpdateRoles, err := c.Store.GetUserRoles(nil, request.Id)
		if err != nil {
			return store.User{}, err
		}

		claims := ctx.Value("claims").(map[string]interface{})
		requesterId := claims["userId"].(string)
		if isAdmin, ok := claims[store.ROLE_ADMIN]; ok && isAdmin.(bool) {
			// ok
		}

		if isOfficeManager, ok := claims[store.ROLE_OFFICE_MANAGER]; ok && isOfficeManager.(bool) {
			// if store.IsAffiliatedTo(userToUpdate.Id, requesterId)
			// ok
		}
	*/

	user, err := c.Store.GetUser(nil, request.Id)
	if err != nil {
		return store.User{}, errors.Wrap(err, "failed to update user")
	}

	for _, role := range roles {
		if !user.Is(role) {
			return store.User{}, errors.Errorf("user %s it not a %s", user.UserId, role)
		}
	}

	if err := c.setAndStoreDecodedImage(ctx, &request); err != nil {
		return store.User{}, err
	}

	user, err = c.Store.UpdateUser(nil, store.User{
		UserId:    dbString(request.Id),
		Email:     dbString(request.Email),
		Address_1: dbString(request.Address_1),
		Address_2: dbString(request.Address_2),
		City:      dbString(request.City),
		State:     dbString(request.State),
		Zip:       dbString(request.Zip),
		Phone:     dbString(request.Phone),
		Gender:    dbString(request.Gender),
		LastName:  dbString(request.LastName),
		FirstName: dbString(request.FirstName),
		ImageUri:  dbString(request.ImageUri),
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
	user, err := c.Store.GetUser(nil, request.Id)
	if err != nil {
		return store.User{}, errors.Wrap(err, "failed to get user")
	}

	for _, role := range roles {
		if !user.Is(role) {
			return store.User{}, errors.Errorf("user %s is not a %s", user.UserId, role)
		}
	}

	c.setBucketUri(ctx, &user)

	return user, nil
}

func (c *UserService) GetUserByEmail(ctx context.Context, request UserTransport) (store.User, error) {
	user, err := c.Store.GetUserByEmail(nil, request.Email)
	if err != nil {
		return store.User{}, errors.Wrap(err, "failed to get user")
	}

	c.setBucketUri(ctx, &user)

	return user, nil
}

func (c *UserService) DeleteUserByRoles(ctx context.Context, request UserTransport, roles ...string) error {
	user, err := c.Store.GetUser(nil, request.Id)
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
	users, err := c.Store.ListUsers(nil, roleConstraint)
	if err != nil {
		return make([]store.User, 0), err
	}

	for i := range users {
		c.setBucketUri(ctx, &users[i])
	}
	return users, nil
}

// ServiceMiddleware is a chainable behavior modifier for adultResponsibleService.
type ServiceMiddleware func(UserService) UserService
