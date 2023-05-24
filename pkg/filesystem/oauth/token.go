package oauth

import "context"

type TokenProvider interface {
	UpdateCredential(ctx context.Context, isSlave bool) error
	AccessToken() string
}
