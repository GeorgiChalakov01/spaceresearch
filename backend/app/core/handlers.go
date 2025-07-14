package core

import (
	"net/http"
)

func WithAuthentication(handler func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
