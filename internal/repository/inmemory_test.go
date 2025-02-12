package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_CreatePost(t *testing.T) {
	storage := NewInMemoryStorage()

	// Создаем пост
	post, err := storage.CreatePost(1, true, "Test Post", "This is a test post.")
	assert.NoError(t, err)
	assert.Equal(t, 2, post.ID) // Первый пост уже есть в NewInMemoryStorage
	assert.Equal(t, "Test Post", post.Title)

	// Пытаемся создать пост с тем же заголовком и содержимым
	_, err = storage.CreatePost(1, true, "Test Post", "This is a test post.")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "post with this title and content already exists")
}

func TestInMemoryStorage_GetAllPosts(t *testing.T) {
	storage := NewInMemoryStorage()

	// Получаем все посты
	posts, err := storage.GetAllPosts()
	assert.NoError(t, err)
	assert.Len(t, posts, 1) // Первый пост уже есть в NewInMemoryStorage
}

func TestInMemoryStorage_GetPostAndComments(t *testing.T) {
	storage := NewInMemoryStorage()

	// Получаем пост и комментарии
	posts, comments, err := storage.GetPostAndComments("First Post")
	assert.NoError(t, err)
	assert.Len(t, posts, 1)
	assert.Len(t, comments, 1)

	// Пытаемся получить несуществующий пост
	_, _, err = storage.GetPostAndComments("Non-existent Post")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "post with title Non-existent Post not found")
}

func TestInMemoryStorage_CreateComment(t *testing.T) {
	storage := NewInMemoryStorage()

	// Создаем комментарий
	comment, err := storage.CreateComment(1, 1, nil, "This is a test comment.")
	assert.NoError(t, err)
	assert.Equal(t, 2, comment.ID) // Первый комментарий уже есть в NewInMemoryStorage

	// Пытаемся создать комментарий к несуществующему посту
	_, err = storage.CreateComment(1, 999, nil, "This is a test comment.")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "post not found")

	// Пытаемся создать комментарий длиннее 2000 символов
	longContent := string(make([]byte, 2001))
	_, err = storage.CreateComment(1, 1, nil, longContent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comment length exceeds 2000 characters")
}

func TestInMemoryStorage_DisableComments(t *testing.T) {
	storage := NewInMemoryStorage()

	// Отключаем комментарии
	err := storage.DisableComments(1, 1, false)
	assert.NoError(t, err)

	// Пытаемся отключить комментарии для несуществующего поста
	err = storage.DisableComments(999, 1, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "post not found")

	// Пытаемся отключить комментарии для поста, где пользователь не автор
	err = storage.DisableComments(1, 2, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only the author can modify comment permissions")
}

func TestInMemoryStorage_RegisterUser(t *testing.T) {
	storage := NewInMemoryStorage()

	// Регистрируем нового пользователя
	user, err := storage.RegisterUser("testuser", "password")
	assert.NoError(t, err)
	assert.Equal(t, 2, user.ID) // Первый пользователь уже есть в NewInMemoryStorage

	// Пытаемся зарегистрировать пользователя с тем же именем
	_, err = storage.RegisterUser("testuser", "password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user with name testuser already exists")
}

func TestInMemoryStorage_LoginUser(t *testing.T) {
	storage := NewInMemoryStorage()

	// Логиним существующего пользователя
	user, err := storage.LoginUser("Lilla", "secret123")
	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)

	// Пытаемся залогинить несуществующего пользователя
	_, err = storage.LoginUser("nonexistent", "password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")

	// Пытаемся залогинить с неправильным паролем
	_, err = storage.LoginUser("Lilla", "wrongpassword")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect password")
}

func TestInMemoryStorage_IsUserRegistered(t *testing.T) {
	storage := NewInMemoryStorage()

	// Проверяем зарегистрированного пользователя
	assert.True(t, storage.IsUserRegistered(1))

	// Проверяем незарегистрированного пользователя
	assert.False(t, storage.IsUserRegistered(999))
}
