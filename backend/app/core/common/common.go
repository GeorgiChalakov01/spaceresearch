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
