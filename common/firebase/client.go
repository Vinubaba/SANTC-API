package firebase

import (
	"context"
	"firebase.google.com/go/auth"
	"github.com/pkg/errors"
)

type Client struct {
	FirebaseClient *auth.Client `inject:""`
}

func (c *Client) DeleteUserByEmail(ctx context.Context, email string) error {
	user, err := c.FirebaseClient.GetUserByEmail(ctx, email)
	if err != nil {
		return errors.Wrap(err, "user not found in firebase")
	}

	return c.FirebaseClient.DeleteUser(ctx, user.UID)
}

func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return c.FirebaseClient.VerifyIDToken(ctx, idToken)
}

func (c *Client) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	return c.FirebaseClient.GetUser(ctx, uid)
}

func (c *Client) SetCustomUserClaims(ctx context.Context, uid string, customClaims map[string]interface{}) error {
	return c.FirebaseClient.SetCustomUserClaims(ctx, uid, customClaims)
}
