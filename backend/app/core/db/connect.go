package db

import (
	"os"
	"context"
	"github.com/jackc/pgx/v5"
)

func Connect() (*pgx.Conn, error) {
	containerName := os.Getenv("POSTGRES_CONTAINER_NAME")
	port := os.Getenv("POSTGRES_PORT")

	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	schema := os.Getenv("POSTGRES_DB")

	url := "postgres://" + user + ":" + pass + "@" + containerName + ":" + port + "/" + schema

	conn, err := pgx.Connect(context.Background(), url)
	return conn, err
}
