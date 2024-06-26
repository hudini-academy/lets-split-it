package main

import (
	"expense/pkg/jwt"
	"net/http"
)

// requireAuthenticatedUser check whether the user is logged in or not.
func (app *Application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.AuthenticatedUser(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
		next.ServeHTTP(w, r)
	})
}

func (app *Application) TokenAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authToken, err := r.Cookie("authToken")
		if err != nil {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		authenticated, err := jwt.ParseToken(authToken.Value)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		if !authenticated {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
