package repository

import (
	"PostComm/internal/models"
	"database/sql"
)

type Storage interface {
	CreatePost(userID int, allowComments bool, title, content string) (models.Post, error)
	GetAllPosts() ([]models.Post, error)
	GetPostAndComments(title string) ([]models.Post, []models.Comment, error)
	CreateComment(userID, postID int, parentID *int, content string) (models.Comment, error)
	GetComments(postID, limit, offset int) ([]models.Comment, error)
	DisableComments(postID, userID int, allowComments bool) error
	RegisterUser(name, password string) (models.User, error)
	LoginUser(name, password string) (models.User, error)
	IsUserRegistered(userID int) bool
}

type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
