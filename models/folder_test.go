package model

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetFolderByPath(t *testing.T) {
	asserts := assert.New(t)

	//policyRows := sqlmock.NewRows([]string{"id", "name"}).
	//	AddRow(1, "默认上传策略")
	//mock.ExpectQuery("^SELECT (.+)").WillReturnRows(policyRows)

	folder,_ := GetFolderByPath("/测试/test",1)
	fmt.Println(folder)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.NoError(mock.)
}
