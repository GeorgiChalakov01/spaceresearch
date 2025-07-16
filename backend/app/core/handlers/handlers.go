package handlers

import (
	"os"
	"fmt"
	"errors"
	"context"
	"net/http"
	"github.com/jackc/pgx/v5"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/common"
)

func sessionCookieValid(conn *pgx.Conn, r *http.Request) error {
	var AuthError = errors.New("Unauthorized")
	cookie, err := r.Cookie("user_email")
        if err != nil {
                return err 
        }
	email := cookie.Value


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
			fmt.Printf("\nCould not take the value of cookie (user_email). Error:\n%v", err)
			// http.Redirect(w, r, "/err?error=emailCookieReading", http.StatusSeeOther)
			http.Redirect(w, r, "/signin?error=authenticationError", http.StatusSeeOther)
			return
		}
		handler(w, r, conn)
	})
}

func WithOutAuthentication(handler func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn)) http.HandlerFunc {
	return WithDB(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
		if err := sessionCookieValid(conn, r); err != nil {
			handler(w, r, conn)
		} else {
			http.Redirect(w, r, "/home?error=authenticationError", http.StatusSeeOther)
			return
		}
	})
}

func WithDB(handler func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		containerName := os.Getenv("POSTGRES_CONTAINER_NAME")
		port := os.Getenv("POSTGRES_PORT")

		user := os.Getenv("POSTGRES_USER")
		pass := os.Getenv("POSTGRES_PASSWORD")
		schema := os.Getenv("POSTGRES_DB")

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
