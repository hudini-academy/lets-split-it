package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

// routes handles the routing of the application.
func (app *Application) routes() http.Handler {
	// Creates a middleware chain.
	middlewareChain := alice.New(app.Session.Enable)
	Auth := alice.New(app.requireAuthenticatedUser)
	mux := pat.New()
	mux.Get("/login", http.HandlerFunc(app.Login))
	mux.Post("/login", http.HandlerFunc(app.LoginUser))
	mux.Get("/", Auth.ThenFunc(app.Home))
	mux.Get("/allusers", Auth.ThenFunc(app.AllUsers))
    mux.Get("/adduser", Auth.ThenFunc(app.AddUserform))
    mux.Post("/adduser", Auth.ThenFunc(app.AddUser))
	mux.Get("/logout", Auth.ThenFunc(app.Logout))
	mux.Get("/submit_expense",Auth.ThenFunc(app.GetAddSplitForm))
	mux.Post("/submit_expense",Auth.ThenFunc(app.AddSplit))
	mux.Get("/expense_details",Auth.ThenFunc(app.ExpenseDetails))
	
	mux.Get("/deleteuser",  Auth.ThenFunc(app.DeleteUser))
	mux.Get("/cancelexpense",  Auth.ThenFunc(app.Cancelexpense))
	mux.Get("/markaspaid",Auth.ThenFunc(app.MarkAsPaid))

	fileServer := http.FileServer(http.Dir(app.Config.StaticDir)) // serve static files
	mux.Get("/static/", http.StripPrefix("/static", fileServer))  // strip static directory.
	mux.Get("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./ui/images"))))

	return middlewareChain.Then(mux)
}
