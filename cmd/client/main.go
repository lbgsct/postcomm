package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const endpoint = "http://localhost:8080/graphql"

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type UserAuth struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

var currentUser *UserAuth = nil

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to the PostComm client!")
	fmt.Println("Type 'help' to see available commands, or 'exit' to quit.")

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		cmd := parts[0]

		switch strings.ToLower(cmd) {
		case "exit", "quit":
			fmt.Println("Exiting...")
			return

		case "help":
			usage()

		case "register":
			args := strings.Split(strings.TrimSpace(partsArg(parts)), " ")
			if len(args) < 2 {
				fmt.Println("Usage: register <username> <password>")
				continue
			}
			name := args[0]
			password := args[1]
			registerUser(name, password)

		case "login":
			args := strings.Split(strings.TrimSpace(partsArg(parts)), " ")
			if len(args) < 2 {
				fmt.Println("Usage: login <username> <password>")
				continue
			}
			name := args[0]
			password := args[1]
			loginUser(name, password)

		case "list":
			listPosts()

		case "post":
			title := partsArg(parts)
			if title == "" {
				fmt.Println("Usage: post <title>")
				continue
			}
			findPost(title)

		case "create-post":
			args := parseArgs(partsArg(parts))
			if len(args) < 3 {
				fmt.Println("Usage: create-post \"<title>\" \"<content>\" <allowcomm>")
				continue
			}

			title := args[0]
			content := args[1]
			allowComm := args[2]

			createPost(title, content, allowComm)

		case "create-comment":
			args := parseArgs(partsArg(parts))
			if len(args) < 2 {
				fmt.Println("Usage: create-comment \"<post_title>\" \"<content>\" [parentId]")
				continue
			}

			postTitle := args[0]
			content := args[1]
			var parentId *string
			if len(args) > 2 {
				parentId = &args[2]
			}

			createComment(postTitle, parentId, content)

		default:
			fmt.Println("Unknown command:", cmd)
			usage()
		}
	}
}

func partsArg(parts []string) string {
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func usage() {
	fmt.Println("Commands:")
	fmt.Println("  help                                     - Показать эту справку")
	fmt.Println("  exit / quit                              - Выйти из программы")
	fmt.Println("  register <username> <password>           - Зарегистрироваться")
	fmt.Println("  login <username> <password>              - Войти")
	fmt.Println("  list                                     - Получить список всех постов")
	fmt.Println("  post <title>                             - Получить пост по названию")
	fmt.Println("  create-post <title> <content> <allowcomm> - Создать новый пост")
	fmt.Println("  create-comment \"<post_title>\" \"<content>\" [parentId] - Создать новый комментарий")
}

func sendGraphQLRequest(query string) (*GraphQLResponse, error) {
	data := map[string]string{"query": query}
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, err
	}
	return &gqlResp, nil
}

func listPosts() {
	query := `query { posts { id title content } }`
	resp, err := sendGraphQLRequest(query)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	printResponse(resp)
}

func findPost(title string) {
	query := fmt.Sprintf(`query { post(title: %q) { post { id title content } comments { id content parentId children { id content } } } }`, title)
	resp, err := sendGraphQLRequest(query)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	printResponse(resp)
}

func registerUser(name, password string) {
	query := fmt.Sprintf(`mutation {
      registerUser(name: %q, password: %q) {
        id
        name
        password
      }
    }`, name, password)
	resp, err := sendGraphQLRequest(query)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	printResponse(resp)
}

func loginUser(name, password string) {
	query := fmt.Sprintf(`mutation {
      login(name: %q, password: %q) {
        user {
          id
          name
        }
        token
      }
    }`, name, password)

	resp, err := sendGraphQLRequest(query)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	printResponse(resp)

	var loginData struct {
		Login struct {
			User struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"user"`
			Token string `json:"token"`
		} `json:"login"`
	}
	if err := json.Unmarshal(resp.Data, &loginData); err == nil {
		currentUser = &UserAuth{
			ID:    loginData.Login.User.ID,
			Name:  loginData.Login.User.Name,
			Token: loginData.Login.Token,
		}
		fmt.Printf("Now logged in as %s (user ID = %d)\n", currentUser.Name, currentUser.ID)
	} else {
		fmt.Println("Error parsing login response:", err)
	}
}

func createPost(title, content, allowCommStr string) {
	if currentUser == nil {
		fmt.Println("You must be logged in to create a post.")
		return
	}

	allowComm := strings.ToLower(strings.TrimSpace(allowCommStr)) == "true"

	query := fmt.Sprintf(`mutation {
		createPost(userId: %d, title: %q, content: %q, allowcomm: %t) {
			id
			title
			content
			allowcomm
		}
	}`, currentUser.ID, title, content, allowComm)

	resp, err := sendGraphQLRequest(query)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	printResponse(resp)
}

func createComment(postTitle string, parentId *string, content string) {
	if currentUser == nil {
		fmt.Println("You must be logged in to create a comment.")
		return
	}

	postID, err := findPostIDByTitle(postTitle)
	if err != nil {
		fmt.Println("Error finding post:", err)
		return
	}

	mutation := ""
	if parentId == nil {
		mutation = fmt.Sprintf(
			`mutation {
                createComment(userId: %d, postId: %d, content: %q) {
                    id postId userId content parentId
                }
            }`,
			currentUser.ID, postID, content)
	} else {
		var pId int
		_, err := fmt.Sscanf(*parentId, "%d", &pId)
		if err != nil {
			fmt.Println("Invalid parentId:", err)
			return
		}
		mutation = fmt.Sprintf(
			`mutation {
                createComment(userId: %d, postId: %d, content: %q, parentId: %d) {
                    id postId userId content parentId
                }
            }`,
			currentUser.ID, postID, content, pId)
	}

	resp, err := sendGraphQLRequest(mutation)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	printResponse(resp)
}

func findPostIDByTitle(title string) (int, error) {
	query := fmt.Sprintf(`query { post(title: %q) { post { id } } }`, title)
	resp, err := sendGraphQLRequest(query)
	if err != nil {
		return 0, fmt.Errorf("failed to find post: %v", err)
	}

	var result struct {
		Post struct {
			Post struct {
				ID int `json:"id"`
			} `json:"post"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %v", err)
	}

	if result.Post.Post.ID == 0 {
		return 0, fmt.Errorf("post not found")
	}

	return result.Post.Post.ID, nil
}

func printResponse(resp *GraphQLResponse) {
	if len(resp.Errors) > 0 {
		fmt.Println("Errors:")
		for _, e := range resp.Errors {
			fmt.Println(" -", e.Message)
		}
		return
	}
	// Красиво форматируем поле data
	var out bytes.Buffer
	if err := json.Indent(&out, resp.Data, "", "  "); err != nil {
		fmt.Println("Response:", string(resp.Data))
	} else {
		fmt.Println("Response:")
		fmt.Println(out.String())
	}
}

func parseArgs(input string) []string {
	re := regexp.MustCompile(`"[^"]*"|\S+`)
	matches := re.FindAllString(input, -1)

	for i, match := range matches {
		matches[i] = strings.Trim(match, `"`)
	}

	return matches
}
