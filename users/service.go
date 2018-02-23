package users

import (
	"context"
	"fmt"
	"strings"

	"github.com/DigitalFrameworksLLC/teddycare/storage"
	"github.com/DigitalFrameworksLLC/teddycare/store"

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

	GetUserRoles(ctx context.Context, request UserTransport) ([]store.Role, error)

	AddPendingUser(ctx context.Context, request UserTransport) error
}

type UserService struct {
	Store interface {
		AddUser(tx *gorm.DB, user store.User) (store.User, error)
		ListUsers(tx *gorm.DB, roleConstraint string) ([]store.User, error)
		UpdateUser(tx *gorm.DB, user store.User) (store.User, error)
		DeleteUser(tx *gorm.DB, userId string) (err error)
		GetUser(tx *gorm.DB, userId string) (store.User, error)

		AddRole(tx *gorm.DB, role store.Role) (store.Role, error)
		GetUserRoles(tx *gorm.DB, userId string) ([]store.Role, error)

		GetPendingConnexionRoles(tx *gorm.DB, email string) ([]store.PendingConnexionRole, error)
		DeletePendingConnexionRole(tx *gorm.DB, role store.PendingConnexionRole) error
		CreatePendingConnexionRole(tx *gorm.DB, role store.PendingConnexionRole) error
		Tx() *gorm.DB
	} `inject:""`
	FirebaseClient interface {
		DeleteUser(ctx context.Context, uid string) error
	} `inject:"teddyFirebaseClient"`
	Storage storage.Storage `inject:""`
}

func (c *UserService) AddUserByRoles(ctx context.Context, request UserTransport, roles ...string) (store.User, error) {
	tx := c.Store.Tx()

	if err := c.setAndStoreDecodedImage(ctx, &request); err != nil {
		return store.User{}, err
	}

	createdUser, err := c.Store.AddUser(tx, store.User{
		Email:     request.Email,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Gender:    request.Gender,
		Zip:       request.Zip,
		State:     request.State,
		Phone:     request.Phone,
		City:      request.City,
		Address_1: request.Address_1,
		Address_2: request.Address_2,
		UserId:    request.Id,
		ImageUri:  request.ImageUri,
	})
	if err != nil {
		tx.Rollback()
		return store.User{}, errors.New("failed to create user")
	}

	for _, role := range roles {
		_, err := c.Store.AddRole(tx, store.Role{
			Role:   role,
			UserId: createdUser.UserId,
		})
		if err != nil {
			tx.Rollback()
			return store.User{}, errors.New("failed to set user role")
		}
	}

	tx.Commit()
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
		UserId:    request.Id,
		Email:     request.Email,
		Address_1: request.Address_1,
		Address_2: request.Address_2,
		City:      request.City,
		State:     request.State,
		Zip:       request.Zip,
		Phone:     request.Phone,
		Gender:    request.Gender,
		LastName:  request.LastName,
		FirstName: request.FirstName,
		ImageUri:  request.ImageUri,
	})
	if err != nil {
		return store.User{}, err
	}

	c.setBucketUri(ctx, &user)
	return user, nil
}

func (c *UserService) setBucketUri(ctx context.Context, user *store.User) {
	if user.ImageUri == "" {
		return
	}
	if !strings.Contains(user.ImageUri, "/") {
		uri, err := c.Storage.Get(ctx, user.ImageUri)
		if err != nil {
			// todo logger
			fmt.Println("failed to generate image uri")
			user.ImageUri = ""
		}
		user.ImageUri = uri
	}
}

func (c *UserService) getRoles(ctx context.Context, user store.User) ([]string, error) {
	roles := make([]string, 0)
	userRoles, err := c.Store.GetUserRoles(nil, user.UserId)
	for _, role := range userRoles {
		roles = append(roles, role.Role)
	}
	return roles, err
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

func (c *UserService) hasRole(roles []store.Role, role string) bool {
	for _, r := range roles {
		if r.Role == role {
			return true
		}
	}
	return false
}

func (c *UserService) GetUserRoles(ctx context.Context, request UserTransport) ([]store.Role, error) {
	return c.Store.GetUserRoles(nil, request.Id)
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

	if err := c.FirebaseClient.DeleteUser(ctx, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete user from firebase")
	}

	if err := c.Store.DeleteUser(nil, request.Id); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if err := c.Storage.Delete(ctx, user.ImageUri); err != nil {
		fmt.Println("failed to delete user image")
		// todo: logger here. Maybe the image uri is not in our buckets, so silently fail
		//return errors.Wrap(err, "failed to delete user image")
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

func (c *UserService) AddPendingUser(ctx context.Context, request UserTransport) error {
	if len(request.Roles) == 0 {
		return errors.New("invalid request: roles are empty")
	}
	for _, role := range request.Roles {
		if err := c.Store.CreatePendingConnexionRole(nil, store.PendingConnexionRole{
			Email: request.Email,
			Role:  role,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (c *UserService) GetPendingConnexionRoles(ctx context.Context, email string) ([]store.PendingConnexionRole, error) {
	return c.Store.GetPendingConnexionRoles(nil, email)
}

func (c *UserService) DeletePendingConnexionRoles(ctx context.Context, pendingRoles []store.PendingConnexionRole) error {
	for _, role := range pendingRoles {
		if err := c.Store.DeletePendingConnexionRole(nil, role); err != nil {
			return err
		}
	}
	return nil
}

// ServiceMiddleware is a chainable behavior modifier for adultResponsibleService.
type ServiceMiddleware func(UserService) UserService
