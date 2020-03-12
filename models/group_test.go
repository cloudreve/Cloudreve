package model

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetGroupByID(t *testing.T) {
	asserts := assert.New(t)

	//找到用户组时
	groupRows := sqlmock.NewRows([]string{"id", "name", "policies"}).
		AddRow(1, "管理员", "[1]")
	mock.ExpectQuery("^SELECT (.+)").WillReturnRows(groupRows)

	group, err := GetGroupByID(1)
	asserts.NoError(err)
	asserts.Equal(Group{
		Model: gorm.Model{
			ID: 1,
		},
		Name:       "管理员",
		Policies:   "[1]",
		PolicyList: []uint{1},
	}, group)

	//未找到用户时
	mock.ExpectQuery("^SELECT (.+)").WillReturnError(errors.New("not found"))
	group, err = GetGroupByID(1)
	asserts.Error(err)
	asserts.Equal(Group{}, group)
}

func TestGroup_AfterFind(t *testing.T) {
	asserts := assert.New(t)

	testCase := Group{
		Model: gorm.Model{
			ID: 1,
		},
		Name:     "管理员",
		Policies: "[1]",
	}
	err := testCase.AfterFind()
	asserts.NoError(err)
	asserts.Equal(testCase.PolicyList, []uint{1})

	testCase.Policies = "[1,2,3,4,5]"
	err = testCase.AfterFind()
	asserts.NoError(err)
	asserts.Equal(testCase.PolicyList, []uint{1, 2, 3, 4, 5})

	testCase.Policies = "[1,2,3,4,5"
	err = testCase.AfterFind()
	asserts.Error(err)

	testCase.Policies = "[]"
	err = testCase.AfterFind()
	asserts.NoError(err)
	asserts.Equal(testCase.PolicyList, []uint{})
}

func TestGroup_BeforeSave(t *testing.T) {
	asserts := assert.New(t)
	group := Group{
		PolicyList: []uint{1, 2, 3},
	}
	{
		err := group.BeforeSave()
		asserts.NoError(err)
		asserts.Equal("[1,2,3]", group.Policies)
	}

}
