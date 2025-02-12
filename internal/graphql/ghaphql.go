package graphql

import (
	"PostComm/internal/models"
	"PostComm/internal/repository"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"
)

var storage repository.Storage

func InitSchema(repo repository.Storage) {
	storage = repo
}

// Тип данных для поста
var postType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Post",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"userId": &graphql.Field{
			Type: graphql.Int,
		},
		"title": &graphql.Field{ // название поста
			Type: graphql.String,
		},
		"content": &graphql.Field{ // содержание
			Type: graphql.String,
		},
		"allowcomm": &graphql.Field{
			Type: graphql.Boolean,
		},
	},
})

// Тип данных для комментария
var commentType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Comment",
	Fields: graphql.Fields{
		"id":       &graphql.Field{Type: graphql.Int},
		"postId":   &graphql.Field{Type: graphql.Int},
		"userId":   &graphql.Field{Type: graphql.Int},
		"content":  &graphql.Field{Type: graphql.String},
		"parentId": &graphql.Field{Type: graphql.Int},
	},
})

func init() {
	commentType.AddFieldConfig("children", &graphql.Field{
		Type: graphql.NewList(commentType),
	})
}

// Тип данных для поста с комментариями
var postWithCommentType = graphql.NewObject(graphql.ObjectConfig{
	Name: "PostWithComment",
	Fields: graphql.Fields{
		"post": &graphql.Field{
			Type: postType,
		},
		"comments": &graphql.Field{
			Type: graphql.NewList(commentType),
		},
	},
})

// Тип данных для регистрации пользователя
var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"password": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// Тип данных для аутентификации
var authUserType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthUser",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type: userType,
		},
		"token": &graphql.Field{
			Type: graphql.String,
		},
	},
})

// Схема
var RootQuery = graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		// Запрос на получение поста/постов по названию
		"post": &graphql.Field{
			Type: postWithCommentType,
			Args: graphql.FieldConfigArgument{
				"title": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				title := p.Args["title"].(string)
				posts, comments, err := storage.GetPostAndComments(title)
				if err != nil {
					return nil, fmt.Errorf("failed to get post and comments: %v", err)
				}
				if len(posts) == 0 {
					return nil, fmt.Errorf("post not found")
				}

				nestedComments := buildCommentHierarchy(comments) // Используем исправленную функцию

				result := map[string]interface{}{
					"post":     posts[0],
					"comments": nestedComments,
				}
				return result, nil
			},
		},

		"posts": &graphql.Field{
			Type: graphql.NewList(postType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				posts, err := storage.GetAllPosts()
				if err != nil {
					return nil, fmt.Errorf("failed to get posts: %v", err)
				}
				return posts, nil
			},
		},
	},
}

// Мутации
var Mutation = graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		// Мутация для создания нового поста
		"createPost": &graphql.Field{
			Type: postType,
			Args: graphql.FieldConfigArgument{
				"userId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"title": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"content": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"allowcomm": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Boolean),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				userId := p.Args["userId"].(int)
				title := p.Args["title"].(string)
				content := p.Args["content"].(string)
				allowcomm := p.Args["allowcomm"].(bool)

				post, err := storage.CreatePost(userId, allowcomm, title, content)
				if err != nil {
					return nil, fmt.Errorf("failed to create post: %v", err)
				}
				return post, nil
			},
		},
		// Мутация для добавления комментария
		"createComment": &graphql.Field{
			Type: commentType,
			Args: graphql.FieldConfigArgument{
				"userId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"postId": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"content": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"parentId": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				userId := p.Args["userId"].(int)
				postId := p.Args["postId"].(int)
				content := p.Args["content"].(string)
				var parentId *int
				if p.Args["parentId"] != nil {
					tmp := p.Args["parentId"].(int)
					parentId = &tmp
				}

				comment, err := storage.CreateComment(userId, postId, parentId, content)
				if err != nil {
					return nil, fmt.Errorf("failed to create comment: %v", err)
				}
				return comment, nil
			},
		},
		// Мутация для создания пользователя
		"registerUser": &graphql.Field{
			Type: userType,
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"password": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				name := p.Args["name"].(string)
				password := p.Args["password"].(string)

				user, err := storage.RegisterUser(name, password)
				if err != nil {
					return nil, fmt.Errorf("failed to register user: %v", err)
				}
				return user, nil
			},
		},
		// Мутация для входа пользователя
		"login": &graphql.Field{
			Type: authUserType,
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"password": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				name := p.Args["name"].(string)
				password := p.Args["password"].(string)

				user, err := storage.LoginUser(name, password)
				if err != nil {
					return nil, fmt.Errorf("failed to login: %v", err)
				}

				token := "dummy-token" // Здесь можно добавить логику генерации токена
				return map[string]interface{}{
					"user":  user,
					"token": token,
				}, nil
			},
		},
		// Мутация для отключения комментариев
		"disableComments": &graphql.Field{
			Type: graphql.Boolean,
			Args: graphql.FieldConfigArgument{
				"postId":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				"userId":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
				"allowcomm": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Boolean)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				postId := p.Args["postId"].(int)
				userId := p.Args["userId"].(int)
				allowcomm := p.Args["allowcomm"].(bool)

				err := storage.DisableComments(postId, userId, allowcomm)
				if err != nil {
					return nil, fmt.Errorf("failed to disable comments: %v", err)
				}
				return true, nil
			},
		},
	},
}

// Схема GraphQL
var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    graphql.NewObject(RootQuery), // Для обработки запросов
	Mutation: graphql.NewObject(Mutation),  // Для обработки мутаций
})

// Обработчик GraphQL запросов
func Handler(w http.ResponseWriter, r *http.Request) {
	var query string

	if r.Method == http.MethodPost {
		// Если POST, то читаем тело запроса
		var body struct {
			Query string `json:"query"`
		}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}
		query = body.Query
	} else {
		// Если GET, читаем параметр "query" из URL
		query = r.URL.Query().Get("query")
	}
	log.Printf("Received query: %s", query)

	// Проверка: если запрос пустой, возвращаем ошибку
	if query == "" && r.Method == http.MethodPost {
		http.Error(w, "Query is missing", http.StatusBadRequest)
		return
	}

	params := graphql.Params{Schema: Schema, RequestString: query}
	result := graphql.Do(params)
	if len(result.Errors) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result.Errors)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Функция для построения иерархии комментариев
func buildCommentHierarchy(comments []models.Comment) []map[string]interface{} {
	commentMap := make(map[int][]models.Comment)

	// Группируем комментарии по parentId
	for _, comment := range comments {
		parentID := 0
		if comment.ParentID != nil {
			parentID = *comment.ParentID
		}
		commentMap[parentID] = append(commentMap[parentID], comment)
	}

	// Рекурсивная функция для построения дерева
	var buildTree func(parentID int) []map[string]interface{}
	buildTree = func(parentID int) []map[string]interface{} {
		var nested []map[string]interface{}
		for _, comment := range commentMap[parentID] {
			nestedComment := map[string]interface{}{
				"id":       comment.ID,
				"postId":   comment.PostID,
				"userId":   comment.UserID,
				"content":  comment.Content,
				"parentId": comment.ParentID,
				"children": buildTree(comment.ID), // Вложенные комментарии
			}
			nested = append(nested, nestedComment)
		}
		return nested
	}

	return buildTree(0) // 0 означает корневой комментарий
}
