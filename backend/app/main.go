package main


import (
	"fmt"
	"net/http"
	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5"
	"github.com/GeorgiChalakov01/spaceresearch/pages/err"
	"github.com/GeorgiChalakov01/spaceresearch/pages/signup"
	"github.com/GeorgiChalakov01/spaceresearch/pages/home"
	"github.com/GeorgiChalakov01/spaceresearch/core/handlers"
	"github.com/GeorgiChalakov01/spaceresearch/core/auth"
)

func main() {
	http.Handle("/", http.RedirectHandler("/home", http.StatusSeeOther))

	http.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request){
		templ.Handler(err.Error()).ServeHTTP(w, r)
	})

	http.HandleFunc("/signup", handlers.WithAuthentication(func(w http.ResponseWriter, r *http.Request){
		templ.Handler(signup.SignUp()).ServeHTTP(w, r)
	}))

	http.HandleFunc("/process-signup", handlers.WithDB(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
		auth.ProcessSignUp(w, r, conn)
	}))

	http.HandleFunc("/home", handlers.WithAuthentication(func(w http.ResponseWriter, r *http.Request){
		templ.Handler(home.Home()).ServeHTTP(w, r)
	}))

	fmt.Println("Serving application on port 8080...")
	http.ListenAndServe(":8080", nil)
}
