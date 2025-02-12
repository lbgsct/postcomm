package repository

import (
	"PostComm/internal/models"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type InMemoryStorage struct {
	posts    []models.Post
	comments []models.Comment
	users    []models.User
	mu       sync.Mutex
}

func NewInMemoryStorage() *InMemoryStorage {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)

	return &InMemoryStorage{
		posts: []models.Post{
			{ID: 1, UserID: 1, Title: "First Post", Content: "This is the first post content.", AllowComments: true},
		},
		comments: []models.Comment{
			{ID: 1, PostID: 1, UserID: 1, Content: "This is the first comment"},
		},
		users: []models.User{
			{ID: 1, Name: "Lilla", Password: string(hashed)},
		},
	}
}

func (s *InMemoryStorage) CreatePost(userID int, allowComments bool, title, content string) (models.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsUserRegistered(userID) {
		return models.Post{}, fmt.Errorf("user with id %d is not registered", userID)
	}

	for _, post := range s.posts {
		if post.Title == title && post.Content == content {
			return models.Post{}, fmt.Errorf("post with this title and content already exists")
		}
	}

	newPost := models.Post{
		ID:            len(s.posts) + 1,
		UserID:        userID,
		Title:         title,
		Content:       content,
		AllowComments: allowComments,
	}
	s.posts = append(s.posts, newPost)
	return newPost, nil
}

func (s *InMemoryStorage) GetAllPosts() ([]models.Post, error) {
	return s.posts, nil
}

func (s *InMemoryStorage) GetPostAndComments(title string) ([]models.Post, []models.Comment, error) {
	var resultPosts []models.Post
	var resultComments []models.Comment

	for _, post := range s.posts {
		if post.Title == title {
			resultPosts = append(resultPosts, post)
			for _, comment := range s.comments {
				if comment.PostID == post.ID {
					resultComments = append(resultComments, comment)
				}
			}
			break
		}
	}

	if len(resultPosts) == 0 {
		return nil, nil, fmt.Errorf("post with title %s not found", title)
	}

	return resultPosts, resultComments, nil
}

func (s *InMemoryStorage) CreateComment(userID, postID int, parentID *int, content string) (models.Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsUserRegistered(userID) {
		return models.Comment{}, fmt.Errorf("user with id %d is not registered", userID)
	}

	var postExists bool
	for _, post := range s.posts {
		if post.ID == postID {
			postExists = true
			if !post.AllowComments {
				return models.Comment{}, errors.New("comments are disabled for this post")
			}
			break
		}
	}

	if !postExists {
		return models.Comment{}, errors.New("post not found")
	}

	if len(content) > 2000 {
		return models.Comment{}, errors.New("comment length exceeds 2000 characters")
	}

	newComment := models.Comment{
		ID:       len(s.comments) + 1,
		PostID:   postID,
		UserID:   userID,
		Content:  content,
		ParentID: parentID,
	}

	s.comments = append(s.comments, newComment)
	return newComment, nil
}

func (s *InMemoryStorage) GetComments(postID, limit, offset int) ([]models.Comment, error) {
	var result []models.Comment
	count := 0

	for _, comment := range s.comments {
		if comment.PostID == postID {
			if count >= offset && count < offset+limit {
				result = append(result, comment)
			}
			count++
		}
	}
	return result, nil
}

func (s *InMemoryStorage) DisableComments(postID, userID int, allowComments bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, post := range s.posts {
		if post.ID == postID {
			if post.UserID != userID {
				return errors.New("only the author can modify comment permissions")
			}
			s.posts[i].AllowComments = allowComments
			return nil
		}
	}
	return errors.New("post not found")
}

func (s *InMemoryStorage) RegisterUser(name, password string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, user := range s.users {
		if user.Name == name {
			return models.User{}, fmt.Errorf("user with name %s already exists", name)
		}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := models.User{
		ID:       len(s.users) + 1,
		Name:     name,
		Password: string(hashed),
	}
	s.users = append(s.users, newUser)
	return newUser, nil
}

func (s *InMemoryStorage) LoginUser(name, password string) (models.User, error) {
	for _, user := range s.users {
		if user.Name == name {
			err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				return models.User{}, errors.New("incorrect password")
			}
			return user, nil
		}
	}
	return models.User{}, errors.New("user not found")
}

func (s *InMemoryStorage) IsUserRegistered(userID int) bool {
	for _, user := range s.users {
		if user.ID == userID {
			return true
		}
	}
	return false
}
