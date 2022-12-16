package middleware

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestValidateSourceLink(t *testing.T) {
	a := assert.New(t)
	rec := httptest.NewRecorder()
	testFunc := ValidateSourceLink()

	// ID 不存在
	{
		c, _ := gin.CreateTestContext(rec)
		testFunc(c)
		a.True(c.IsAborted())
	}

	// SourceLink 不存在
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set("object_id", 1)
		mock.ExpectQuery("SELECT(.+)source_links(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
		testFunc(c)
		a.True(c.IsAborted())
		a.NoError(mock.ExpectationsWereMet())
	}

	// 原文件不存在
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set("object_id", 1)
		mock.ExpectQuery("SELECT(.+)source_links(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery("SELECT(.+)files(.+)").WithArgs(0).WillReturnRows(sqlmock.NewRows([]string{"id"}))
		testFunc(c)
		a.True(c.IsAborted())
		a.NoError(mock.ExpectationsWereMet())
	}

	// 成功
	{
		c, _ := gin.CreateTestContext(rec)
		c.Set("object_id", 1)
		mock.ExpectQuery("SELECT(.+)source_links(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "file_id"}).AddRow(1, 2))
		mock.ExpectQuery("SELECT(.+)files(.+)").WithArgs(2).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE(.+)source_links").WillReturnResult(sqlmock.NewResult(1, 1))
		testFunc(c)
		a.False(c.IsAborted())
		a.NoError(mock.ExpectationsWereMet())
	}

}
