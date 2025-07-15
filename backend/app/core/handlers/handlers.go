package handlers

import (
	"os"
	"fmt"
	"errors"
	"context"
	"net/http"
	"github.com/jackc/pgx/v5"
	"github.com/GeorgiChalakov01/spaceresearch/core/common"
)

func sessionCookieValid(conn *pgx.Conn, r *http.Request) error {
	var AuthError = errors.New("Unauthorized")
	email := common.GetCookieValue(r, "user_email")

	user, err := common.GetUserData(conn, email)
	if err != nil {
		return AuthError
	}

	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken.Value == "" || sessionToken.Value != user.SessionToken {
		return AuthError
	}

	return nil
}


func WithAuthentication(handler func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn)) http.HandlerFunc {
	return WithDB(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
		if err := sessionCookieValid(conn, r); err != nil {
			fmt.Println(err)
			http.Redirect(w, r, "/signin?error=authenticationError", http.StatusSeeOther)
			return
		}
		handler(w, r, conn)
	})
}

func WithOutAuthentication(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return WithDB(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
		if err := sessionCookieValid(conn, r); err == nil {
			http.Redirect(w, r, "/home?error=authenticationError", http.StatusSeeOther)
			return
		}
		handler(w, r)
	})
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
