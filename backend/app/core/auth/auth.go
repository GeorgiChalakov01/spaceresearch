package auth

import (
	"fmt"
	"context"
	"net/http"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/GeorgiChalakov01/spaceresearch/core/common"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}



func createUser (conn *pgx.Conn, user common.User) error {
	userCount, err := common.CountUsers(conn)
	if err != nil {
		fmt.Println("Could not count the users. Error: ")
		return err
	}
	if userCount == 0 {
		fmt.Println("User will be created as an admin")
		user.IsAdmin = true
	} else {
		user.IsAdmin = false
	}

	// Start a transaction
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
	INSERT INTO users (name, email, passwordHash, sessionToken, csrfToken, isAdmin ) VALUES ($1, $2, $3, $4, $5, $6)`, user.Name, user.Email, user.PasswordHash, user.SessionToken, user.CSRFToken, user.IsAdmin)

	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func ProcessSignUp(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	user := common.User {
		Email: r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	var err error
	user.PasswordHash, err = hashPassword(user.Password)
	if err != nil {
		fmt.Printf("Couldn't hash password.")
		http.RedirectHandler("/err?error=hashingError", http.StatusSeeOther)
	}

	if err := common.GenerateAndSetTokens(w, &user); err != nil {
		http.Redirect(w, r, "/err?error=tokenGenerationFailed", http.StatusSeeOther)
		return
	}

	if err := createUser(conn, user); err != nil {
		fmt.Printf("Couldn't create a user.")
		http.RedirectHandler("/err?error=createUserError", http.StatusSeeOther)
	}
}
