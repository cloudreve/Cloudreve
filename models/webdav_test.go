package model

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetWebdavByPassword(t *testing.T) {
	asserts := assert.New(t)
	mock.ExpectQuery("SELECT(.+)").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	_, err := GetWebdavByPassword("e", 1)
	asserts.NoError(mock.ExpectationsWereMet())
	asserts.Error(err)
}
