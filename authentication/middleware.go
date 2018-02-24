package authentication

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	. "github.com/DigitalFrameworksLLC/teddycare/shared"
	"github.com/DigitalFrameworksLLC/teddycare/store"
	"github.com/DigitalFrameworksLLC/teddycare/users"

	"firebase.google.com/go/auth"
)

type Authenticator struct {
	FirebaseClient interface {
		VerifyIDToken(idToken string) (*auth.Token, error)
		GetUser(ctx context.Context, uid string) (*auth.UserRecord, error)
		SetCustomUserClaims(ctx context.Context, uid string, customClaims map[string]interface{}) error
	} `inject:""`
	UserService interface {
		GetPendingConnexionRoles(ctx context.Context, email string) ([]store.PendingConnexionRole, error)
		DeletePendingConnexionRoles(ctx context.Context, pendingRoles []store.PendingConnexionRole) error
		AddUserByRoles(ctx context.Context, request users.UserTransport, roles ...string) (store.User, error)
	} `inject:""`
	Logger *Logger `inject:""`
}

func (f *Authenticator) Roles(next http.Handler, roles ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		claims := req.Context().Value("claims").(map[string]interface{})
		if !f.hasRole(roles, claims) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (f *Authenticator) Firebase(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
		token, err := f.FirebaseClient.VerifyIDToken(bearerToken[1])
		if err != nil {
			HttpError(w, NewError(fmt.Sprintf("invalid authorization token: %s", err.Error())), http.StatusBadRequest)
			return
		}

		// Lookup the user associated with the specified uid.
		firebaseUser, err := f.FirebaseClient.GetUser(req.Context(), token.UID)
		if err != nil {
			HttpError(w, NewError(fmt.Sprintf("failed to retrieve user from firebase: %s", err.Error())), http.StatusBadRequest)
			return
		}

		if !f.hasAtLeastOneRoleInCustomClaim(firebaseUser.CustomClaims) {
			// checker email office manager pending
			pendingRoles, err := f.UserService.GetPendingConnexionRoles(nil, firebaseUser.Email)
			if err != nil {
				HttpError(w, NewError(fmt.Sprintf("failed to get pending roles: %s", err.Error())), http.StatusInternalServerError)
				return
			}

			roles := make([]string, 0)
			for _, r := range pendingRoles {
				roles = append(roles, r.Role)
			}
			if len(roles) == 0 {
				roles = append(roles, ROLE_ADULT)
			}
			if _, err = f.UserService.AddUserByRoles(ctx, users.UserTransport{
				Id:        firebaseUser.UID,
				Email:     firebaseUser.Email,
				FirstName: firebaseUser.DisplayName,
				LastName:  "",
				ImageUri:  firebaseUser.PhotoURL,
				Phone:     firebaseUser.PhoneNumber,
			}, roles...); err != nil {
				HttpError(w, NewError(fmt.Sprintf("failed to add user: %s", err.Error())), http.StatusInternalServerError)
				return
			}

			claims := map[string]interface{}{
				"userId":            firebaseUser.UID,
				ROLE_TEACHER:        false,
				ROLE_OFFICE_MANAGER: false,
				ROLE_ADULT:          false,
				ROLE_ADMIN:          false,
			}
			for _, role := range roles {
				claims[role] = true
			}
			if err = f.FirebaseClient.SetCustomUserClaims(ctx, firebaseUser.UID, claims); err != nil {
				HttpError(w, NewError(err.Error()), http.StatusInternalServerError)
				return
			}
			firebaseUser.CustomClaims = claims
			ctx = context.WithValue(ctx, "claims", claims)
			if err := f.UserService.DeletePendingConnexionRoles(ctx, pendingRoles); err != nil {
				f.Logger.Warn(ctx, "fail to remove pending roles...", "err", err.Error())
			}
		}

		req = req.WithContext(context.WithValue(context.Background(), "claims", firebaseUser.CustomClaims))

		next.ServeHTTP(w, req)
	})
}

func (f *Authenticator) hasAtLeastOneRoleInCustomClaim(claims map[string]interface{}) bool {
	if isAdult, ok := claims[ROLE_ADULT]; ok && isAdult.(bool) {
		return true
	}
	if isOfficeManager, ok := claims[ROLE_OFFICE_MANAGER]; ok && isOfficeManager.(bool) {
		return true
	}
	if isAdmin, ok := claims[ROLE_ADMIN]; ok && isAdmin.(bool) {
		return true
	}
	if isTeacher, ok := claims[ROLE_TEACHER]; ok && isTeacher.(bool) {
		return true
	}
	return false
}

func (f *Authenticator) hasRole(roles []string, customClaim map[string]interface{}) bool {
	for _, role := range roles {
		if r, ok := customClaim[role]; ok {
			if r.(bool) {
				return true
			}
		}
	}
	return false
}
