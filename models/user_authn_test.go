package model

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUser_RegisterAuthn(t *testing.T) {
	asserts := assert.New(t)
	credential := webauthn.Credential{}
	user := User{
		Model: gorm.Model{ID: 1},
	}

	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		user.RegisterAuthn(&credential)
		asserts.NoError(mock.ExpectationsWereMet())
	}
}

func TestUser_WebAuthnCredentials(t *testing.T) {
	asserts := assert.New(t)
	user := User{
		Model: gorm.Model{ID: 1},
		Authn: `[{"ID":"123","PublicKey":"+4sg1vYcjg/+=","AttestationType":"packed","Authenticator":{"AAGUID":"+lg==","SignCount":0,"CloneWarning":false}}]`,
	}
	{
		credentials := user.WebAuthnCredentials()
		asserts.Len(credentials, 1)
	}
}

func TestUser_WebAuthnDisplayName(t *testing.T) {
	asserts := assert.New(t)
	user := User{
		Model: gorm.Model{ID: 1},
		Nick:  "123",
	}
	{
		nick := user.WebAuthnDisplayName()
		asserts.Equal("123", nick)
	}
}

func TestUser_WebAuthnIcon(t *testing.T) {
	asserts := assert.New(t)
	user := User{
		Model: gorm.Model{ID: 1},
	}
	{
		icon := user.WebAuthnIcon()
		asserts.NotEmpty(icon)
	}
}

func TestUser_WebAuthnID(t *testing.T) {
	asserts := assert.New(t)
	user := User{
		Model: gorm.Model{ID: 1},
	}
	{
		id := user.WebAuthnID()
		asserts.Len(id, 8)
	}
}

func TestUser_WebAuthnName(t *testing.T) {
	asserts := assert.New(t)
	user := User{
		Model: gorm.Model{ID: 1},
		Email: "abslant@foxmail.com",
	}
	{
		name := user.WebAuthnName()
		asserts.Equal("abslant@foxmail.com", name)
	}
}

func TestUser_RemoveAuthn(t *testing.T) {
	asserts := assert.New(t)
	user := User{
		Model: gorm.Model{ID: 1},
		Authn: `[{"ID":"123","PublicKey":"+4sg1vYcjg/+=","AttestationType":"packed","Authenticator":{"AAGUID":"+lg==","SignCount":0,"CloneWarning":false}}]`,
	}
	{
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		user.RemoveAuthn("123")
		asserts.NoError(mock.ExpectationsWereMet())
	}
}
