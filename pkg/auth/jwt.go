package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudreve/Cloudreve/v4/ent"
	"github.com/cloudreve/Cloudreve/v4/inventory"
	"github.com/cloudreve/Cloudreve/v4/pkg/hashid"
	"github.com/cloudreve/Cloudreve/v4/pkg/logging"
	"github.com/cloudreve/Cloudreve/v4/pkg/serializer"
	"github.com/cloudreve/Cloudreve/v4/pkg/setting"
	"github.com/cloudreve/Cloudreve/v4/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type TokenAuth interface {
	// Issue issues a new pair of credentials for the given user.
	Issue(ctx context.Context, u *ent.User) (*Token, error)
	// VerifyAndRetrieveUser verifies the given token and inject the user into current context.
	// Returns if upper caller should continue process other session provider.
	VerifyAndRetrieveUser(c *gin.Context) (bool, error)
	// Refresh refreshes the given refresh token and returns a new pair of credentials.
	Refresh(ctx context.Context, refreshToken string) (*Token, error)
}

// Token stores token pair for authentication
type Token struct {
	AccessToken    string    `json:"access_token"`
	RefreshToken   string    `json:"refresh_token"`
	AccessExpires  time.Time `json:"access_expires"`
	RefreshExpires time.Time `json:"refresh_expires"`

	UID int `json:"-"`
}

type (
	TokenType         string
	TokenIDContextKey struct{}
)

var (
	TokenTypeAccess  = TokenType("access")
	TokenTypeRefresh = TokenType("refresh")

	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrUserNotFound        = errors.New("user not found")
)

const (
	AuthorizationHeader = "Authorization"
	TokenHeaderPrefix   = "Bearer "
)

type Claims struct {
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
	StateHash []byte `json:"state_hash,omitempty"`
}

// NewTokenAuth creates a new token based auth provider.
func NewTokenAuth(idEncoder hashid.Encoder, s setting.Provider, secret []byte, userClient inventory.UserClient, l logging.Logger) TokenAuth {
	return &tokenAuth{
		idEncoder:  idEncoder,
		s:          s,
		secret:     secret,
		userClient: userClient,
		l:          l,
	}
}

type tokenAuth struct {
	l          logging.Logger
	idEncoder  hashid.Encoder
	s          setting.Provider
	secret     []byte
	userClient inventory.UserClient
}

func (t *tokenAuth) Refresh(ctx context.Context, refreshToken string) (*Token, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return t.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.TokenType != TokenTypeRefresh {
		return nil, ErrInvalidRefreshToken
	}

	uid, err := t.idEncoder.Decode(claims.Subject, hashid.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	expectedUser, err := t.userClient.GetActiveByID(ctx, uid)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check if user changed password or revoked session
	expectedHash := t.hashUserState(ctx, expectedUser)
	if !bytes.Equal(claims.StateHash, expectedHash[:]) {
		return nil, ErrInvalidRefreshToken
	}

	return t.Issue(ctx, expectedUser)
}

func (t *tokenAuth) VerifyAndRetrieveUser(c *gin.Context) (bool, error) {
	headerVal := c.GetHeader(AuthorizationHeader)
	if strings.HasPrefix(headerVal, TokenHeaderPrefixCr) {
		// This is an HMAC auth header, skip JWT verification
		return false, nil
	}

	tokenString := strings.TrimPrefix(headerVal, TokenHeaderPrefix)
	if tokenString == "" {
		return true, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return t.secret, nil
	})

	if err != nil {
		t.l.Warning("Failed to parse jwt token: %s", err)
		return false, nil
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.TokenType != TokenTypeAccess {
		return false, serializer.NewError(serializer.CodeCredentialInvalid, "Invalid token type", nil)
	}

	uid, err := t.idEncoder.Decode(claims.Subject, hashid.UserID)
	if err != nil {
		return false, serializer.NewError(serializer.CodeNotFound, "User not found", err)
	}

	util.WithValue(c, inventory.UserIDCtx{}, uid)
	return false, nil
}

func (t *tokenAuth) Issue(ctx context.Context, u *ent.User) (*Token, error) {
	uidEncoded := hashid.EncodeUserID(t.idEncoder, u.ID)
	tokenSettings := t.s.TokenAuth(ctx)
	issueDate := time.Now()
	accessTokenExpired := time.Now().Add(tokenSettings.AccessTokenTTL)
	refreshTokenExpired := time.Now().Add(tokenSettings.RefreshTokenTTL)

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uidEncoded,
			NotBefore: jwt.NewNumericDate(issueDate),
			ExpiresAt: jwt.NewNumericDate(accessTokenExpired),
		},
	}).SignedString(t.secret)
	if err != nil {
		return nil, fmt.Errorf("faield to sign access token: %w", err)
	}

	userHash := t.hashUserState(ctx, u)
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uidEncoded,
			NotBefore: jwt.NewNumericDate(issueDate),
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpired),
		},
		StateHash: userHash[:],
	}).SignedString(t.secret)
	if err != nil {
		return nil, fmt.Errorf("faield to sign refresh token: %w", err)
	}

	return &Token{
		AccessToken:    accessToken,
		RefreshToken:   refreshToken,
		AccessExpires:  accessTokenExpired,
		RefreshExpires: refreshTokenExpired,
		UID:            u.ID,
	}, nil
}

// hashUserState returns a hash string for user state for critical fields, it is used
// to detect refresh token revocation after user changed password.
func (t *tokenAuth) hashUserState(ctx context.Context, u *ent.User) [32]byte {
	return sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s", u.Email, u.Password, t.s.SiteBasic(ctx).ID)))
}
