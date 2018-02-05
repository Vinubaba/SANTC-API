package firebase

import (
	"context"
	"firebase.google.com/go/auth"
)

type Client struct {
	FirebaseClient *auth.Client `inject:""`
}

func (c *Client) DeleteUser(ctx context.Context, uid string) error {
	return c.FirebaseClient.DeleteUser(ctx, uid)
}

func (c *Client) VerifyIDToken(idToken string) (*auth.Token, error) {
	return c.FirebaseClient.VerifyIDToken(idToken)
}

func (c *Client) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	return c.FirebaseClient.GetUser(ctx, uid)
}

func (c *Client) SetCustomUserClaims(ctx context.Context, uid string, customClaims map[string]interface{}) error {
	return c.FirebaseClient.SetCustomUserClaims(ctx, uid, customClaims)
}
