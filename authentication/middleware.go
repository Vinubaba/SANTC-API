package authentication

import (
	"context"
	"net/http"
	"strings"

	. "github.com/DigitalFrameworksLLC/teddycare/shared"

	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702

/*func ValidateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")

		if authorizationHeader == "" {
			HttpError(w, NewError("invalid authorization token"), http.StatusBadRequest)
			return
		}

		bearerToken := strings.Split(authorizationHeader, " ")
		if len(bearerToken) == 2 {
			token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("error token method")
				}
				return []byte(secret), nil
			})
			if err != nil {
				HttpError(w, NewError(err.Error()), http.StatusBadRequest)
				return
			}
			if token.Valid {
				req = req.WithContext(context.WithValue(context.Background(), "decoded", token.Claims))
				next.ServeHTTP(w, req)
			} else {
				HttpError(w, NewError("invalid authorization token"), http.StatusBadRequest)
			}
		}
	})
}*/

func Roles(next http.Handler, roles ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorizationHeader := req.Header.Get("authorization")

		if authorizationHeader == "" {
			HttpError(w, NewError("invalid authorization token"), http.StatusBadRequest)
			return
		}

		bearerToken := strings.Split(authorizationHeader, " ")
		if len(bearerToken) == 2 {
			token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("error token method")
				}
				return []byte(secret), nil
			})
			if err != nil {
				HttpError(w, NewError(err.Error()), http.StatusBadRequest)
				return
			}
			if !token.Valid {
				HttpError(w, NewError("invalid authorization token"), http.StatusBadRequest)
				return
			}

			var claim TeddyClaims
			mapstructure.Decode(token.Claims.(jwt.MapClaims), &claim)

			if !intersects(claim.Roles, roles) {
				HttpError(w, NewError(fmt.Sprintf("you must be %v to use this service", roles)), http.StatusBadRequest)
				return
			}

			req = req.WithContext(context.WithValue(context.Background(), "decoded", token.Claims))

			next.ServeHTTP(w, req)
		}
	})
}

func intersects(list1, list2 []string) bool {
	for _, v1 := range list1 {
		for _, v2 := range list2 {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}
