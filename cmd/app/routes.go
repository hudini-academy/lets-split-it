package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/justinas/alice"
)

// routes sets up and returns an http.Handler that handles all the routes for the application.
func (app *Application) routes() http.Handler {
	// Creates a middleware chain.
	middlewareChain := alice.New(app.Session.Enable)
	Auth := alice.New(app.requireAuthenticatedUser)

	// Initialize a new request multiplexer using pat.
	mux := pat.New()

	mux.Get("/", Auth.ThenFunc(app.Home)) // GET / - Home page after authentication.

	// Routes for handling user authentication and login/logout.
	mux.Get("/login", http.HandlerFunc(app.Login))      // GET /login - Renders login page.
	mux.Post("/login", http.HandlerFunc(app.LoginUser)) // POST /login - Handles user login.
	mux.Get("/logout", Auth.ThenFunc(app.Logout))       // GET /logout - Logs out the user.

	// Routes for user management.
	mux.Get("/allusers", Auth.ThenFunc(app.AllUsers))     // GET /allusers - Displays all users.
	mux.Get("/adduser", Auth.ThenFunc(app.AddUserform))   // GET /adduser - Renders add user form.
	mux.Post("/adduser", Auth.ThenFunc(app.AddUser))      // POST /adduser - Handles user addition.
	mux.Get("/deleteuser", Auth.ThenFunc(app.DeleteUser)) // GET /deleteuser - Renders delete user page.

	// Routes for handling expenses and splits.
	mux.Get("/submit_expense", Auth.ThenFunc(app.GetAddSplitForm)) // GET /submit_expense - Renders expense submission form.
	mux.Post("/submit_expense", Auth.ThenFunc(app.AddSplit))       // POST /submit_expense - Handles expense submission.
	mux.Get("/expense_details", Auth.ThenFunc(app.ExpenseDetails)) // GET /expense_details - Renders expense details page.
	mux.Get("/cancelexpense", Auth.ThenFunc(app.Cancelexpense))    // GET /cancelexpense - Renders cancel expense page.
	mux.Get("/markaspaid", Auth.ThenFunc(app.MarkAsPaid))          // GET /markaspaid - Renders mark as paid page.
	mux.Get("/allsplits", Auth.ThenFunc(app.Allsplits))            // GET /allsplits - Displays all splits.
	mux.Get("/changePassword", Auth.ThenFunc(app.ChangePasswordForm)) // GET /changepassword - Render change password page.
	mux.Post("/changePassword", Auth.ThenFunc(app.ChangePassword)) //POST /changepassword - Changes the user password.

	fileServer := http.FileServer(http.Dir(app.Config.StaticDir))                               // serve static files from configured directory.
	mux.Get("/static/", http.StripPrefix("/static", fileServer))                                // Route to serve static files.
	mux.Get("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./ui/images")))) // Route to serve images.

	return middlewareChain.Then(mux)
}
