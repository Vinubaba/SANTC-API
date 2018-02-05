package authentication

import (
	"context"
	"github.com/DigitalFrameworksLLC/teddycare/store"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"time"
)

const (
	secret = "123456789"
)

type Service interface {
	Authenticate(ctx context.Context, request AuthenticateTransport) (JwtToken, error)
}

type AuthenticationService struct {
	Store interface {
		CheckUserCredentials(tx *gorm.DB, user store.User) (store.User, error)
		GetUserRoles(tx *gorm.DB, userId string) ([]store.Role, error)
	} `inject:""`
}

type JwtToken struct {
	Token string `json:"token"`
}

type TeddyClaims struct {
	UserId string   `json:"userId"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
	jwt.StandardClaims
}

func (s *AuthenticationService) Authenticate(ctx context.Context, request AuthenticateTransport) (JwtToken, error) {
	user, err := s.Store.CheckUserCredentials(nil, store.User{
		Email:    request.Email,
		Password: request.Password,
	})
	if err != nil {
		return JwtToken{}, errors.Wrap(err, "login failed")
	}

	roles, err := s.Store.GetUserRoles(nil, user.UserId)
	if err != nil {
		return JwtToken{}, errors.Wrap(err, "failed to find user roles")
	}

	strRoles := []string{}
	for _, role := range roles {
		strRoles = append(strRoles, role.Role)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, TeddyClaims{
		UserId: user.UserId,
		Email:  user.Email,
		Roles:  strRoles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().UTC().Unix() + 60*60*6, // 6 hours validity
			IssuedAt:  time.Now().UTC().Unix(),
		},
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return JwtToken{}, err
	}
	return JwtToken{Token: tokenString}, nil
}
