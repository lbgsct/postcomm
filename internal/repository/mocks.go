package repository

import (
	"PostComm/internal/models"
	"database/sql"

	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	callArgs := m.Called(query, args)
	return callArgs.Get(0).(sql.Result), callArgs.Error(1)
}

func (m *MockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	callArgs := m.Called(query, args)
	return callArgs.Get(0).(*sql.Rows), callArgs.Error(1)
}

func (m *MockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	row := m.Called(query, args)
	if row.Get(0) == nil {
		return nil
	}
	return row.Get(0).(*sql.Row)
}

// CreatePost - мок для метода CreatePost
func (m *MockDB) CreatePost(userID int, allowComments bool, title, content string) (models.Post, error) {
	args := m.Called(userID, allowComments, title, content)
	return args.Get(0).(models.Post), args.Error(1)
}

func (m *MockDB) GetAllPosts() ([]models.Post, error) {
	args := m.Called()
	return args.Get(0).([]models.Post), args.Error(1)
}

func (m *MockDB) GetPostAndComments(title string) ([]models.Post, []models.Comment, error) {
	args := m.Called(title)
	return args.Get(0).([]models.Post), args.Get(1).([]models.Comment), args.Error(2)
}

func (m *MockDB) CreateComment(userID, postID int, parentID *int, content string) (models.Comment, error) {
	args := m.Called(userID, postID, parentID, content)
	return args.Get(0).(models.Comment), args.Error(1)
}

func (m *MockDB) DisableComments(postID, userID int, allowComments bool) error {
	args := m.Called(postID, userID, allowComments)
	return args.Error(0)
}

func (m *MockDB) RegisterUser(name, password string) (models.User, error) {
	args := m.Called(name, password)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockDB) LoginUser(name, password string) (models.User, error) {
	args := m.Called(name, password)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockDB) IsUserRegistered(userID int) bool {
	args := m.Called(userID)
	return args.Bool(0)
}
