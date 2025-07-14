package handlers

import (
	"os"
	"fmt"
	"context"
	"net/http"
	"github.com/jackc/pgx/v5"
)

func WithAuthentication(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}
}

func WithOutAuthentication(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}
}

func WithDB(handler func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		containerName := os.Getenv("DB_CONTAINER_NAME")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PWD")
		schema := os.Getenv("DB_SCHEMA")
		port := os.Getenv("DB_PORT")

		url := "postgres://" + user + ":" + pass + "@" + containerName + ":" + port + "/" + schema

		conn, err := pgx.Connect(context.Background(), url)

		if err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/err?error=databaseError", http.StatusSeeOther)
			return
		}
		defer conn.Close(context.Background())

		handler(w, r, conn)
	}
}
