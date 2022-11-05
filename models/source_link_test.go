package model

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSourceLink_Link(t *testing.T) {
	a := assert.New(t)
	s := &SourceLink{}
	s.ID = 1

	// 失败
	{
		s.File.Name = string([]byte{0x7f})
		res, err := s.Link()
		a.Error(err)
		a.Empty(res)
	}

	// 成功
	{
		s.File.Name = "filename"
		res, err := s.Link()
		a.NoError(err)
		a.Contains(res, s.Name)
	}
}

func TestGetSourceLinkByID(t *testing.T) {
	a := assert.New(t)
	mock.ExpectQuery("SELECT(.+)source_links(.+)").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "file_id"}).AddRow(1, 2))
	mock.ExpectQuery("SELECT(.+)files(.+)").WithArgs(2).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	res, err := GetSourceLinkByID(1)
	a.NoError(err)
	a.NotNil(res)
	a.EqualValues(2, res.File.ID)
	a.NoError(mock.ExpectationsWereMet())
}

func TestSourceLink_Downloaded(t *testing.T) {
	a := assert.New(t)
	s := &SourceLink{}
	s.ID = 1
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE(.+)source_links(.+)").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	s.Downloaded()
	a.NoError(mock.ExpectationsWereMet())
}
