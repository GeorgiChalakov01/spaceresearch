package auth

import (
	"fmt"
	"time"
	"context"
	"net/http"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/GeorgiChalakov01/spaceresearch/core/common"
	"github.com/GeorgiChalakov01/spaceresearch/core/validation"
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
		http.Redirect(w, r, "/err?error=hashingError", http.StatusSeeOther)
		return
	}

	if err := common.GenerateAndSetTokens(w, &user); err != nil {
		http.Redirect(w, r, "/err?error=tokenGenerationFailed", http.StatusSeeOther)
		return
	}

	if err := createUser(conn, user); err != nil {
		fmt.Printf("Couldn't create a user.")
		http.Redirect(w, r, "/err?error=createUserError", http.StatusSeeOther)
		return
	}

	fmt.Printf("Successfully created account for %s and set cookies.\n", user.Email)
	http.Redirect(w, r, "/home?success=accountCreated", http.StatusSeeOther)
	return
}

func ProcessSignIn(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	user := common.User{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	if err := validation.ValidateEmail(user.Email); err != nil {
		http.Redirect(w, r, "/signin?error=emailNotValid", http.StatusSeeOther)
		return
	}

	userDB, err := common.GetUserData(conn, user.Email)
	if err != nil {
		http.Redirect(w, r, "/signin?error=emailNotFound", http.StatusSeeOther)
		return
	}

	if !checkPasswordHash(user.Password, userDB.PasswordHash) {
		http.Redirect(w, r, "/signin?error=wrongPassword", http.StatusSeeOther)
		return
	}

	if err := common.GenerateAndSetTokens(w, &user); err != nil {
		http.Redirect(w, r, "/signin?error=tokenGenerationFailed", http.StatusSeeOther)
		return
	}

	if err := common.UpdateUserTokens(conn, user); err != nil {
		http.Redirect(w, r, "/signin?error=tokenUpdateFailed", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/home?success=welcomeBack", http.StatusSeeOther)
	return
}

func ProcessSignOut(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	var user common.User	
        cookie, err := r.Cookie("user_email")
        if err != nil {
                fmt.Printf("Could not take the value of cookie (user_email). Error:\n%v", err)
		http.Redirect(w, r, "/err?error=emailCookieReading", http.StatusSeeOther)
                return
        }
	user.Email = cookie.Value

	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now().Add(-time.Hour),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "csrf_token",
		Value:   "",
		Expires: time.Now().Add(-time.Hour),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "user_email",
		Value:   "",
		Expires: time.Now().Add(-time.Hour),
	})

	// Clear tokens from DB
	emptyUser := common.User{
		Email:        user.Email,
		SessionToken: "",
		CSRFToken:    "",
	}
	if err := common.UpdateUserTokens(conn, emptyUser); err != nil {
		http.Redirect(w, r, "/signin?error=tokenClearFailed", http.StatusSeeOther)
		return
	}
	
	http.Redirect(w, r, "/signin?success=signedOut", http.StatusSeeOther)
}
