package main

import (
	"log"
	"net/http"
)

// Home renders the home page of the application.
func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/home.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Retrieve splits where the user is the creator
	s, err := app.Expense.GetYourSplit(app.Session.GetInt(r, "userId"))
	if err != nil {
		app.ErrorLog.Println(err)
		log.Println(err)
	}

	// Retrieve splits where the user is involved
	involvedSplits, err := app.Expense.GetInvolvedSplits(app.Session.GetInt(r, "userId"))
	if err != nil {
		app.ErrorLog.Println(err)
		log.Println(err)
	}

	app.render(w, files, &templateData{
		UserId:        app.Session.GetInt(r, "userId"),
		TitleUserName: app.Session.GetString(r, "userName"),
		YourSplit:     s,
		Involved:      involvedSplits,
	})
}

// Login renders the login page.
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/login.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	app.render(w, files, &templateData{
		Flash: app.Session.PopString(r, "flash"),
	})
}

// LoginUser handles user authentication.
func (app *Application) LoginUser(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/login.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	user := r.FormValue("email")
	password := r.FormValue("password")

	// Validate email and password fields
	if app.Validate(r, user, "email") || app.Validate(r, password, "Ispassword") {
		flash := app.Session.PopString(r, "flash")
		app.render(w, files, &templateData{
			Flash: flash,
			Email: user,
		})
		return
	}

	// Check if email exists in the database
	exists, err := app.User.CheckEmail(user)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	// Authenticate user with email and password
	if exists {
		id, userName, errAuth := app.User.Autenticate(user, password)
		if errAuth != nil {
			app.Session.Put(r, "flash", "Password is incorrect")
			flash := app.Session.PopString(r, "flash")
			app.render(w, files, &templateData{
				Flash: flash,
				Email: user,
			})
			app.ErrorLog.Println(errAuth)
			log.Println(errAuth)
			return
		}
		// Set session variables and redirect on successful authentication
		app.Session.Put(r, "userId", id)
		app.Session.Put(r, "userName", userName)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		app.Session.Put(r, "flash", "Email is incorrect")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// AddUserform renders the add user form.
func (app *Application) AddUserform(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/adduser.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Render the page with flash message and username
	app.render(w, files, &templateData{
		Flash:         app.Session.PopString(r, "flash"),
		TitleUserName: app.Session.GetString(r, "userName"),
	})
}

// AddUser handles adding a new user to the database.
func (app *Application) AddUser(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/adduser.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// If there are validation errors, render the template with errors
	if app.Validate(r, username, "name") || app.Validate(r, email, "email") || app.Validate(r, password, "password") {
		flash := app.Session.PopString(r, "flash")
		app.render(w, files, &templateData{
			Flash:         flash,
			Email:         email,
			Username:      username,
			TitleUserName: app.Session.GetString(r, "userName"),
		})
		return
	}

	// Check if username already exists
	if !app.isValidUser(r, username) {
		app.render(w, files, &templateData{
			Email:         email,
			Flash:         app.Session.PopString(r, "flash"),
			TitleUserName: app.Session.GetString(r, "userName"),
		})
		return
	}

	// Check if email is valid and does not already exist
	if !app.isValidEmail(r, email) {
		app.render(w, files, &templateData{
			Username:      username,
			Flash:         app.Session.PopString(r, "flash"),
			TitleUserName: app.Session.GetString(r, "userName"),
		})
		return
	}

	// Insert user into the database
	err := app.User.InsertUser(username, email, password)
	if err != nil {
		app.ErrorLog.Println("Error inserting user:", err)
		return
	}

	http.Redirect(w, r, "/allusers", http.StatusSeeOther)
}

// Logout handles user logout by removing session data and redirecting to login page.
func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Remove(r, "userId")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// AllUsers retrieves and renders a list of all users.
func (app *Application) AllUsers(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/allusers.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Retrieve list of all users
	userlist, err := app.User.ListUsers()
	if err != nil {
		app.ErrorLog.Println(err.Error())
		log.Println("AllUsers(): ", err)
		return
	}

	app.render(w, files, &templateData{
		UserList:      userlist,
		Flash:         app.Session.PopString(r, "flash"),
		TitleUserName: app.Session.GetString(r, "userName"),
	})
}
