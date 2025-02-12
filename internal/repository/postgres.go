package repository

import (
	"PostComm/internal/models"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	DB DB
}

func NewPostgresDB(connStr string) (*PostgresDB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Connected to PostgreSQL")
	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) CreatePost(userID int, allowComments bool, title, content string) (models.Post, error) {
	query := `INSERT INTO posts (user_id, title, content, allow_comments) 
              VALUES ($1, $2, $3, $4) 
              RETURNING id, user_id, title, content, allow_comments`

	var post models.Post
	err := p.DB.QueryRow(query, userID, title, content, allowComments).Scan(
		&post.ID, &post.UserID, &post.Title, &post.Content, &post.AllowComments,
	)
	if err != nil {
		return models.Post{}, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

func (p *PostgresDB) GetAllPosts() ([]models.Post, error) {
	query := `SELECT id, user_id, title, content, allow_comments FROM posts`
	rows, err := p.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.AllowComments); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return posts, nil
}

func (p *PostgresDB) GetComments(postID, limit, offset int) ([]models.Comment, error) {
	query := `SELECT id, post_id, user_id, content, parent_id FROM comments WHERE post_id = $1 ORDER BY id LIMIT $2 OFFSET $3`
	rows, err := p.DB.Query(query, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.ParentID); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return comments, nil
}

func (p *PostgresDB) GetPostAndComments(title string) ([]models.Post, []models.Comment, error) {
	queryPost := `SELECT id, user_id, title, content, allow_comments FROM posts WHERE title = $1`
	var post models.Post
	err := p.DB.QueryRow(queryPost, title).Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.AllowComments)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("post with title %s not found", title)
		}
		return nil, nil, fmt.Errorf("failed to fetch post: %w", err)
	}

	queryComments := `SELECT id, post_id, user_id, content, parent_id FROM comments WHERE post_id = $1 ORDER BY parent_id NULLS FIRST, id`
	rows, err := p.DB.Query(queryComments, post.ID)
	if err != nil {
		return []models.Post{post}, nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.ParentID); err != nil {
			return []models.Post{post}, nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return []models.Post{post}, nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return []models.Post{post}, comments, nil
}

func (p *PostgresDB) CreateComment(userID, postID int, parentID *int, content string) (models.Comment, error) {
	query := `INSERT INTO comments (post_id, user_id, content, parent_id) VALUES ($1, $2, $3, $4) RETURNING id`
	var commentID int
	err := p.DB.QueryRow(query, postID, userID, content, parentID).Scan(&commentID)
	if err != nil {
		return models.Comment{}, fmt.Errorf("failed to create comment: %w", err)
	}

	return models.Comment{
		ID:       commentID,
		PostID:   postID,
		UserID:   userID,
		Content:  content,
		ParentID: parentID,
	}, nil
}

func (p *PostgresDB) DisableComments(postID, userID int, allowComments bool) error {
	query := `UPDATE posts SET allow_comments = $1 WHERE id = $2 AND user_id = $3`
	res, err := p.DB.Exec(query, allowComments, postID, userID)
	if err != nil {
		return fmt.Errorf("failed to disable comments: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("post not found or user is not the author")
	}
	return nil
}

func (p *PostgresDB) RegisterUser(name, password string) (models.User, error) {
	query := `INSERT INTO users (name, password) VALUES ($1, $2) RETURNING id`
	var userID int
	err := p.DB.QueryRow(query, name, password).Scan(&userID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to register user: %w", err)
	}

	return models.User{
		ID:       userID,
		Name:     name,
		Password: password,
	}, nil
}

func (p *PostgresDB) LoginUser(name, password string) (models.User, error) {
	query := `SELECT id, password FROM users WHERE name = $1`
	var user models.User
	err := p.DB.QueryRow(query, name).Scan(&user.ID, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, fmt.Errorf("user not found")
		}
		return models.User{}, fmt.Errorf("failed to login user: %w", err)
	}

	if user.Password != password {
		return models.User{}, fmt.Errorf("incorrect password")
	}

	return user, nil
}

func (p *PostgresDB) IsUserRegistered(userID int) bool {
	query := `SELECT COUNT(*) FROM users WHERE id = $1`
	var count int
	err := p.DB.QueryRow(query, userID).Scan(&count)
	return err == nil && count > 0
}
