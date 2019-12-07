package model

import (
	"encoding/json"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUserByID(t *testing.T) {
	asserts := assert.New(t)

	//找到用户时
	userRows := sqlmock.NewRows([]string{"id", "deleted_at", "email", "options", "group_id"}).
		AddRow(1, nil, "admin@cloudreve.org", "{}", 1)
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(userRows)

	groupRows := sqlmock.NewRows([]string{"id", "name", "policies"}).
		AddRow(1, "管理员", "[1]")
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(groupRows)

	policyRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "默认存储策略")
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(policyRows)

	user, err := GetUserByID(1)
	asserts.NoError(err)
	asserts.Equal(User{
		Model: gorm.Model{
			ID:        1,
			DeletedAt: nil,
		},
		Email:   "admin@cloudreve.org",
		Options: "{}",
		GroupID: 1,
		Group: Group{
			Model: gorm.Model{
				ID: 1,
			},
			Name:       "管理员",
			Policies:   "[1]",
			PolicyList: []uint{1},
		},
		Policy: Policy{
			Model: gorm.Model{
				ID: 1,
			},
			OptionsSerialized: PolicyOption{
				FileType: []string{},
			},
			Name: "默认存储策略",
		},
	}, user)

	//未找到用户时
	mock.ExpectQuery("^SELECT (.+)").WillReturnError(errors.New("not found"))
	user, err = GetUserByID(1)
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
	asserts.NotEmpty(newUser.OptionsSerialized)
}

func TestUser_AfterFind(t *testing.T) {
	asserts := assert.New(t)

	policyRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "默认存储策略")
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(policyRows)

	newUser := NewUser()
	err := newUser.AfterFind()
	err = newUser.BeforeSave()
	expected := UserOption{}
	err = json.Unmarshal([]byte(newUser.Options), &expected)

	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Equal(expected, newUser.OptionsSerialized)
	asserts.Equal("默认存储策略", newUser.Policy.Name)
}

func TestUser_BeforeSave(t *testing.T) {
	asserts := assert.New(t)

	newUser := NewUser()
	err := newUser.BeforeSave()
	expected, err := json.Marshal(newUser.OptionsSerialized)

	asserts.NoError(err)
	asserts.Equal(string(expected), newUser.Options)
}

func TestUser_GetPolicyID(t *testing.T) {
	asserts := assert.New(t)

	newUser := NewUser()

	testCases := []struct {
		preferred uint
		available []uint
		expected  uint
	}{
		{
			available: []uint{1},
			expected:  1,
		},
		{
			available: []uint{5, 2, 3},
			expected:  5,
		},
		{
			preferred: 1,
			available: []uint{5, 1, 3},
			expected:  1,
		},
		{
			preferred: 9,
			available: []uint{5, 1, 3},
			expected:  5,
		},
	}

	for key, testCase := range testCases {
		newUser.OptionsSerialized.PreferredPolicy = testCase.preferred
		newUser.Group.PolicyList = testCase.available
		asserts.Equal(testCase.expected, newUser.GetPolicyID(), "测试用例 #%d 未通过", key)
	}
}

func TestUser_GetRemainingCapacity(t *testing.T) {
	asserts := assert.New(t)
	newUser := NewUser()

	newUser.Group.MaxStorage = 100
	asserts.Equal(uint64(100), newUser.GetRemainingCapacity())

	newUser.Group.MaxStorage = 100
	newUser.Storage = 1
	asserts.Equal(uint64(99), newUser.GetRemainingCapacity())

	newUser.Group.MaxStorage = 100
	newUser.Storage = 100
	asserts.Equal(uint64(0), newUser.GetRemainingCapacity())

	newUser.Group.MaxStorage = 100
	newUser.Storage = 200
	asserts.Equal(uint64(0), newUser.GetRemainingCapacity())
}

func TestUser_DeductionCapacity(t *testing.T) {
	asserts := assert.New(t)

	userRows := sqlmock.NewRows([]string{"id", "deleted_at", "storage", "options", "group_id"}).
		AddRow(1, nil, 0, "{}", 1)
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(userRows)
	groupRows := sqlmock.NewRows([]string{"id", "name", "policies"}).
		AddRow(1, "管理员", "[1]")
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(groupRows)

	policyRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "默认存储策略")
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(policyRows)

	newUser, err := GetUserByID(1)
	newUser.Group.MaxStorage = 100
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())

	asserts.Equal(false, newUser.IncreaseStorage(101))
	asserts.Equal(uint64(0), newUser.Storage)

	asserts.Equal(true, newUser.IncreaseStorage(1))
	asserts.Equal(uint64(1), newUser.Storage)

	asserts.Equal(true, newUser.IncreaseStorage(99))
	asserts.Equal(uint64(100), newUser.Storage)

	asserts.Equal(false, newUser.IncreaseStorage(1))
	asserts.Equal(uint64(100), newUser.Storage)

	asserts.True(newUser.IncreaseStorage(0))
}

func TestUser_DeductionStorage(t *testing.T) {
	asserts := assert.New(t)

	// 减少零
	{
		user := User{Storage: 1}
		asserts.True(user.DeductionStorage(0))
		asserts.Equal(uint64(1), user.Storage)
	}
	// 正常
	{
		user := User{
			Model:   gorm.Model{ID: 1},
			Storage: 10,
		}
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs(5, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		asserts.True(user.DeductionStorage(5))
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(uint64(5), user.Storage)
	}

	// 减少的超出可用的
	{
		user := User{
			Model:   gorm.Model{ID: 1},
			Storage: 10,
		}
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs(0, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		asserts.False(user.DeductionStorage(20))
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Equal(uint64(0), user.Storage)
	}
}

func TestUser_IncreaseStorageWithoutCheck(t *testing.T) {
	asserts := assert.New(t)

	// 增加零
	{
		user := User{}
		user.IncreaseStorageWithoutCheck(0)
		asserts.Equal(uint64(0), user.Storage)
	}

	// 减少零
	{
		user := User{
			Model: gorm.Model{ID: 1},
		}
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)").WithArgs(10, 1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		user.IncreaseStorageWithoutCheck(10)
		asserts.Equal(uint64(10), user.Storage)
	}
}

func TestGetUserByEmail(t *testing.T) {
	asserts := assert.New(t)

	mock.ExpectQuery("SELECT(.+)").WithArgs("abslant@foxmail.com").WillReturnRows(sqlmock.NewRows([]string{"id", "email"}))
	_, err := GetUserByEmail("abslant@foxmail.com")

	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Error(err)
}

func TestUser_AfterCreate(t *testing.T) {
	asserts := assert.New(t)
	user := User{Model: gorm.Model{ID: 1}}
	mock.ExpectBegin()
	mock.ExpectExec("INSERT(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := user.AfterCreate(DB)
	asserts.NoError(err)
	asserts.NoError(mock.ExpectationsWereMet())
}

func TestUser_Root(t *testing.T) {
	asserts := assert.New(t)
	user := User{Model: gorm.Model{ID: 1}}

	// 根目录存在
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "根目录"))
		root, err := user.Root()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.NoError(err)
		asserts.Equal("根目录", root.Name)
	}

	// 根目录不存在
	{
		mock.ExpectQuery("SELECT(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
		_, err := user.Root()
		asserts.NoError(mock.ExpectationsWereMet())
		asserts.Error(err)
	}
}
