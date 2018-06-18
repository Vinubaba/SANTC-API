package authentication

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	. "github.com/Vinubaba/SANTC-API/api/shared"
	"github.com/Vinubaba/SANTC-API/api/users"
	"github.com/Vinubaba/SANTC-API/common/store"

	"firebase.google.com/go/auth"
	"github.com/Vinubaba/SANTC-API/common/log"
	"github.com/Vinubaba/SANTC-API/common/roles"
)

type Authenticator struct {
	FirebaseClient interface {
		VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
		GetUser(ctx context.Context, uid string) (*auth.UserRecord, error)
		SetCustomUserClaims(ctx context.Context, uid string, customClaims map[string]interface{}) error
	} `inject:""`
	UserService interface {
		GetUserByEmail(ctx context.Context, request users.UserTransport) (store.User, error)
	} `inject:""`
	Logger *log.Logger `inject:""`
}

func (f *Authenticator) Roles(next http.Handler, roles ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		claims := req.Context().Value("claims").(map[string]interface{})
		if !f.isService(roles, req) && !f.hasRole(roles, claims) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (f *Authenticator) Firebase(next http.Handler, excludePath []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Some route are public (users does not need to be authenticated)
		for _, path := range excludePath {
			if req.RequestURI == path {
				next.ServeHTTP(w, req)
				return
			}
		}

		ctx := req.Context()
		authorizationHeader := req.Header.Get("authorization")

		if authorizationHeader == "" {
			HttpError(w, NewError("invalid authorization token"), http.StatusBadRequest)
			return
		}

		bearerToken := strings.Split(authorizationHeader, " ")
		if len(bearerToken) != 2 {
			HttpError(w, NewError("invalid authorization token"), http.StatusBadRequest)
			return
		}

		token, err := f.FirebaseClient.VerifyIDToken(ctx, bearerToken[1])
		if err != nil {
			HttpError(w, NewError(fmt.Sprintf("invalid authorization token: %s", err.Error())), http.StatusBadRequest)
			return
		}

		// Lookup the user associated with the specified uid.
		firebaseUser, err := f.FirebaseClient.GetUser(ctx, token.UID)
		if err != nil {
			HttpError(w, NewError(fmt.Sprintf("failed to retrieve user from firebase: %s", err.Error())), http.StatusBadRequest)
			return
		}

		if !f.hasAtLeastOneRoleInCustomClaim(firebaseUser.CustomClaims) {
			// lookup database user with email
			user, err := f.UserService.GetUserByEmail(ctx, users.UserTransport{Email: firebaseUser.Email})
			if err != nil {
				HttpError(w, NewError(fmt.Sprintf("user not registered: %s", err.Error())), http.StatusForbidden)
				return
			}

			claims := map[string]interface{}{
				"userId":                  user.UserId.String,
				"daycareId":               user.DaycareId.String,
				roles.ROLE_TEACHER:        false,
				roles.ROLE_OFFICE_MANAGER: false,
				roles.ROLE_ADULT:          false,
				roles.ROLE_ADMIN:          false,
			}
			for _, role := range user.Roles.ToList() {
				claims[role] = true
			}
			if err = f.FirebaseClient.SetCustomUserClaims(ctx, firebaseUser.UID, claims); err != nil {
				HttpError(w, NewError(err.Error()), http.StatusInternalServerError)
				return
			}

			firebaseUser.CustomClaims = claims
			ctx = context.WithValue(ctx, "claims", claims)
		}

		req = req.WithContext(context.WithValue(ctx, "claims", firebaseUser.CustomClaims))
		next.ServeHTTP(w, req)
	})
}

func (f *Authenticator) hasAtLeastOneRoleInCustomClaim(claims map[string]interface{}) bool {
	if isAdult, ok := claims[roles.ROLE_ADULT]; ok && isAdult.(bool) {
		return true
	}
	if isOfficeManager, ok := claims[roles.ROLE_OFFICE_MANAGER]; ok && isOfficeManager.(bool) {
		return true
	}
	if isAdmin, ok := claims[roles.ROLE_ADMIN]; ok && isAdmin.(bool) {
		return true
	}
	if isTeacher, ok := claims[roles.ROLE_TEACHER]; ok && isTeacher.(bool) {
		return true
	}
	return false
}

func (f *Authenticator) hasRole(listRoles []string, customClaim map[string]interface{}) bool {
	for _, role := range listRoles {
		if r, ok := customClaim[role]; ok {
			if r.(bool) {
				return true
			}
		}
	}
	return false
}

func (f *Authenticator) isService(listRoles []string, r *http.Request) bool {
	for _, role := range listRoles {
		if role == roles.ROLE_SERVICE {
			if r.Header.Get(roles.ROLE_REQUEST_HEADER) == roles.ROLE_SERVICE {
				return true
			}
		}
	}
	return false
}
