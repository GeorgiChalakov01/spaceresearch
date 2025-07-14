package main


import (
	"net/http"
	"github.com/a-h/templ"
	"github.com/GeorgiChalakov01/spaceresearch/pages/home"
)

func main() {
	http.Handle("/", http.RedirectHandler("/home", http.StatusSeeOther))
	http.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		templ.Handler(home.Home()).ServeHTTP(w, r)
	})

	http.ListenAndServe(":8080", nil)
}
