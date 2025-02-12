package main

import (
	"PostComm/internal/graphql"
	"PostComm/internal/repository"
	"fmt"
	"log"
	"net/http"
	"os"
)

var storage repository.Storage

func main() {
	storageType := os.Getenv("STORAGE_TYPE") // "memory" или "postgres"

	switch storageType {
	case "postgres":
		db, err := repository.NewPostgresDB("postgres://user:password@localhost/postcomm?sslmode=disable")
		if err != nil {
			log.Fatal("Failed to connect to PostgreSQL:", err)
		}
		storage = db
		fmt.Println("Using PostgreSQL storage")
	default:
		storage = repository.NewInMemoryStorage()
		fmt.Println("Using in-memory storage")
	}

	graphql.InitSchema(storage)

	http.HandleFunc("/graphql", graphql.Handler)
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
