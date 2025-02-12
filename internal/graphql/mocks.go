package graphql

import (
	"PostComm/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockStorage реализует интерфейс repository.Storage
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetAllPosts() ([]models.Post, error) {
	args := m.Called()
	return args.Get(0).([]models.Post), args.Error(1)
}

func (m *MockStorage) GetPostAndComments(title string) ([]models.Post, []models.Comment, error) {
	args := m.Called(title)
	return args.Get(0).([]models.Post), args.Get(1).([]models.Comment), args.Error(2)
}

func (m *MockStorage) CreatePost(userID int, allowComments bool, title, content string) (models.Post, error) {
	args := m.Called(userID, allowComments, title, content)
	return args.Get(0).(models.Post), args.Error(1)
}

func (m *MockStorage) CreateComment(userID, postID int, parentID *int, content string) (models.Comment, error) {
	args := m.Called(userID, postID, parentID, content)
	return args.Get(0).(models.Comment), args.Error(1)
}

func (m *MockStorage) RegisterUser(name, password string) (models.User, error) {
	args := m.Called(name, password)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockStorage) LoginUser(name, password string) (models.User, error) {
	args := m.Called(name, password)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockStorage) DisableComments(postID, userID int, allowComments bool) error {
	args := m.Called(postID, userID, allowComments)
	return args.Error(0)
}
