package main


import (
	"os"
	"fmt"
	"net/http"
	"github.com/a-h/templ"
	"github.com/jackc/pgx/v5"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/pages/err"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/pages/signup"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/pages/signin"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/pages/home"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/pages/upload"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/handlers"
	"github.com/GeorgiChalakov01/spaceresearch/backend/app/core/auth"
	coreUpload "github.com/GeorgiChalakov01/spaceresearch/backend/app/core/upload"
)

func main() {
	http.Handle("/", http.RedirectHandler("/home", http.StatusSeeOther))

	http.HandleFunc("/err", func(w http.ResponseWriter, r *http.Request){
		templ.Handler(err.Error()).ServeHTTP(w, r)
	})

	http.HandleFunc("/signup", handlers.WithOutAuthentication(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
		templ.Handler(signup.SignUp()).ServeHTTP(w, r)
	}))
	http.HandleFunc("/process-signup", handlers.WithOutAuthentication(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
		auth.ProcessSignUp(w, r, conn)
	}))

	http.HandleFunc("/signin", handlers.WithOutAuthentication(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
		templ.Handler(signin.SignIn()).ServeHTTP(w, r)
	}))
	http.HandleFunc("/process-signin", handlers.WithOutAuthentication(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
		auth.ProcessSignIn(w, r, conn)
	}))

	http.HandleFunc("/signout", handlers.WithAuthentication(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
		auth.ProcessSignOut(w, r, conn)
	}))

	http.HandleFunc("/home", handlers.WithAuthentication(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
		templ.Handler(home.Home()).ServeHTTP(w, r)
	}))

	http.HandleFunc("/upload", handlers.WithAuthentication(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
		templ.Handler(upload.Upload()).ServeHTTP(w, r)
	}))
	http.HandleFunc("/process-upload", handlers.WithAuthentication(func(w http.ResponseWriter, r *http.Request, conn *pgx.Conn){
		coreUpload.ProcessUpload(w, r, conn)
	}))

	port := os.Getenv("BACKEND_PORT")
	fmt.Printf("Serving application on port %s ...\n", port)
	http.ListenAndServe(":" + port, nil)
}
