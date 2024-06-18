package main

import (
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
