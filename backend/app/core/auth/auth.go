package auth

import (
	"fmt"
	"net/http"
	"github.com/jackc/pgx/v5"
)

func ProcessSignUp(w http.ResponseWriter, r *http.Request, conn *pgx.Conn) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	fmt.Println(email)
	fmt.Println(password)
}
