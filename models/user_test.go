package model

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUser(t *testing.T) {
	asserts := assert.New(t)

	// 准备数据库 Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	DB, _ = gorm.Open("mysql", db)
	defer db.Close()

	//找到用户时
	rows := sqlmock.NewRows([]string{"id", "deleted_at", "email"}).
		AddRow(1, nil, "admin@cloudreve.org")

	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(rows)

	user, err := GetUser(1)
	asserts.NoError(err)
	asserts.Equal(User{
		Model: gorm.Model{
			ID:        1,
			DeletedAt: nil,
		},
		Email: "admin@cloudreve.org",
	}, user)

	//未找到用户时
	mock.ExpectQuery("^SELECT (.+)").WillReturnError(errors.New("not found"))
	user, err = GetUser(1)
	asserts.Error(err)
	asserts.Equal(User{}, user)
}

func TestUser_SetPassword(t *testing.T) {
	asserts := assert.New(t)
	user := User{}
	err := user.SetPassword("Cause Sega does what nintendon't")
	asserts.NoError(err)
	asserts.NotEmpty(user.Password)
}

func TestUser_CheckPassword(t *testing.T) {
	asserts := assert.New(t)
	user := User{}
	err := user.SetPassword("Cause Sega does what nintendon't")
	asserts.NoError(err)

	//密码正确
	res, err := user.CheckPassword("Cause Sega does what nintendon't")
	asserts.NoError(err)
	asserts.True(res)

	//密码错误
	res, err = user.CheckPassword("Cause Sega does what Nintendon't")
	asserts.NoError(err)
	asserts.False(res)

	//密码字段为空
	user = User{}
	res, err = user.CheckPassword("Cause Sega does what nintendon't")
	asserts.Error(err)
	asserts.False(res)

}

func TestNewUser(t *testing.T) {
	asserts := assert.New(t)
	newUser := NewUser()
	asserts.IsType(User{}, newUser)
	asserts.NotEmpty(newUser.Avatar)
	asserts.NotEmpty(newUser.Options)
}
