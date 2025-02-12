package models

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Post struct {
	ID            int    `json:"id"`
	UserID        int    `json:"user_id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	AllowComments bool   `json:"allow_comments"`
}

type Comment struct {
	ID       int    `json:"id"`
	PostID   int    `json:"post_id"`
	UserID   int    `json:"user_id"`
	Content  string `json:"content"`
	ParentID *int   `json:"parent_id"`
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func CheckPassword(hashedPassword, password string) error {
	if hashedPassword == "" || password == "" {
		return errors.New("password and hashed password cannot be empty")
	}
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
