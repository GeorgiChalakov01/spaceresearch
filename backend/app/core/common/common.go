package common

import (
	"fmt"
	"time"
	"context"
	"net/http"
	"math/rand"
	"encoding/base64"
	"github.com/jackc/pgx/v5"
)

type User struct {
	Id int
	Name string
	Email string
	Password string
	RepeatedPassword string  
	PasswordHash string
	SessionToken string
	CSRFToken string
	IsAdmin bool
}

func CountUsers(conn *pgx.Conn) (int, error) {
	rows, err := conn.Query(context.Background(), "SELECT count(id) FROM users")
	if err != nil {
		return -1, err
	}

	var count int
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return -1, err
		}
	}
	return count, nil
}

func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func GenerateAndSetTokens(w http.ResponseWriter, user *User) error {
	var err error
	user.SessionToken, err = GenerateToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate session token: %w", err)
	}

	user.CSRFToken, err = GenerateToken(32)
	if err != nil {
		return fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    user.SessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    user.CSRFToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "user_email",
		Value:    user.Email,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	return nil
}

func GetCookieValue(r *http.Request, cookieName string) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		fmt.Printf("Could not take the value of cookie (%s). %v", cookieName, err)
		return ""
	}
	return cookie.Value
}

func UpdateUserTokens(conn *pgx.Conn, user User) error {
	// Start a transaction
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "UPDATE users SET sessionToken = $1, csrfToken = $2 WHERE email = $3", user.SessionToken, user.CSRFToken, user.Email)

	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func GetUserData(conn *pgx.Conn, email string) (User, error) {
	var user User
	err := conn.QueryRow(
		context.Background(),
		"SELECT id, name, email, passwordHash, sessionToken, csrfToken, isAdmin FROM users WHERE email=$1", email).Scan(
			&user.Id, &user.Name, &user.Email, &user.PasswordHash, &user.SessionToken, &user.CSRFToken, &user.IsAdmin)
	if err != nil {
		return user, err
	}
	return user, nil
}
