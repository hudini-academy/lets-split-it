package main

import (
	"fmt"
	"log"
	"net/http"

	"strconv"
	"strings"
	"unicode/utf8"
)

// Home page for the application.
func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/home.page.tmpl",
		"ui/html/base.layout.tmpl",
	}
	s, err := app.Expense.GetYourSplit(app.Session.GetInt(r, "userId"))
	involvedSplits, errInvolved := app.Expense.GetInvolvedSplits(app.Session.GetInt(r, "userId"))
	if errInvolved != nil {
		app.ErrorLog.Println()
		log.Println(errInvolved)
	}

	if err != nil {
		app.ErrorLog.Println()
		log.Println(err)
	}
	app.render(w, files, &templateData{
		UserId:    app.Session.GetInt(r, "userId"),
		YourSplit: s,
		Involved:  involvedSplits,
	})

}

// Login shows the logins page.
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/login.page.tmpl",
		"ui/html/base.layout.tmpl",
	}
	app.render(w, files, &templateData{
		Flash: app.Session.PopString(r, "flash"),
	})
}

// Login the user after authentication.
func (app *Application) LoginUser(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/login.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	user := r.FormValue("email")
	password := r.FormValue("password")

	if app.Validate(r, user, "email") || app.Validate(r, password, "Ispassword") {
		flash := app.Session.PopString(r, "flash")
		app.render(w, files, &templateData{
			Flash: flash,
			Email: user,
		})
		return
	}
	exists, err := app.User.CheckEmail(user)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}
	if exists {
		id, errAuth := app.User.Autenticate(user, password)
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
		app.Session.Put(r, "userId", id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		app.Session.Put(r, "flash", "Email is incorrect")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

}

// AddUserForm renders the add user form.
func (app *Application) AddUserform(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/adduser.page.tmpl",
		"ui/html/base.layout.tmpl",
	}
	app.render(w, files, &templateData{
		Flash: app.Session.PopString(r, "flash"),
	})
}

// AddUser adds a new user to the database.
func (app *Application) AddUser(w http.ResponseWriter, r *http.Request) {

	files := []string{
		"ui/html/adduser.page.tmpl",
		"ui/html/base.layout.tmpl",
	}
	// Retrieve form values
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	app.Validate(r, username, "name")
	app.Validate(r, email, "email")
	app.Validate(r, password, "password")
	flash := app.Session.PopString(r, "flash")
	log.Println("Passed")
	// If there are validation errors, render the template with errors
	if app.Validate(r, username, "name") || app.Validate(r, email, "email") || app.Validate(r, password, "password") {
		app.render(w, files, &templateData{
			Flash:    flash,
			Email:    email,
			Username: username,
		})
		return
	}

	// Check if username already exists
	if !app.isValidUser(r, username) {
		app.render(w, files, &templateData{
			Email: email,
			Flash: app.Session.PopString(r, "flash"),
		})
		return
	}

	// Check if email is valid and does not already exist
	if !app.isValidEmail(r, email) {
		app.render(w, files, &templateData{
			Username: username,
			Flash:    app.Session.PopString(r, "flash"),
		})
		return
	}

	// Insert user into the database
	err := app.User.InsertUser(username, email, password)
	if err != nil {
		app.ErrorLog.Println("Error inserting user:", err)
		return
	}

	// Redirect to /adduser on success (to clear form)
	http.Redirect(w, r, "/adduser", http.StatusSeeOther)
}

func (app *Application) Validate(r *http.Request, field string, fieldType string) bool {
	switch fieldType {
	case "name":
		if strings.TrimSpace(field) == "" {
			app.Session.Put(r, "flash", "The name field is blank!")
			return true
		} else if utf8.RuneCountInString(field) > 100 {
			app.Session.Put(r, "flash", "The name field is too long (maximum is 100 characters)!")
			return true
		}
	case "email":
		if strings.TrimSpace(field) == "" {
			app.Session.Put(r, "flash", "The email field is blank!")
			return true
		}
	case "password":
		if utf8.RuneCountInString(field) < 8 {
			app.Session.Put(r, "flash", "The password is too short (minimum is 8 characters)!")
			return true
		}
	case "Ispassword":
		if strings.TrimSpace(field) == "" {
			app.Session.Put(r, "flash", "The password field is blank!")
			return true
		}
	}
	return false
}

// Logout functionality.
func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Remove(r, "userId")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// AllUsers get the list of all users.
func (app *Application) AllUsers(w http.ResponseWriter, r *http.Request) {
	userlist, err := app.User.ListUsers()
	if err != nil {
		app.ErrorLog.Println(err.Error())
		log.Println("AllUsers(): ", err)
		return
	}

	files := []string{
		"ui/html/allusers.page.tmpl",
		"ui/html/base.layout.tmpl",
	}
	app.render(w, files, &templateData{
		UserList: userlist,
	})
}

func (app *Application) AddSplit(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	amount := r.FormValue("amount")
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		app.Session.Put(r, "flash", "Invalid amount!")
		return
	}
	note := r.FormValue("note")

	usersSelected := r.Form["user[]"]

	if len(usersSelected) == 0 {
		app.Session.Put(r, "flash", "No users selected!")
		return
	}

	result, err := app.Expense.Insert(note, amountFloat, app.Session.GetInt(r, "userId"))
	if err != nil {
		app.ErrorLog.Fatal()
		return
	}
	app.Session.Put(r, "flash", "Task successfully created!")

	fmt.Println("Ids selected:")
	for _, id := range usersSelected {
		fmt.Println(id)
	}
	expenseId, err := result.LastInsertId()
	if err != nil {
		app.ErrorLog.Fatal()
	}
	log.Println("done.....")
	app.Expense.Insert2Split(expenseId, amountFloat, usersSelected, app.Session.GetInt(r, "userId"))
	http.Redirect(w, r, "/submit_expense", http.StatusSeeOther)

}

func (app *Application) GetAddSplitForm(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/split.page.tmpl",
		"ui/html/base.layout.tmpl",
	}
	userList, errGettingList := app.User.GetAllUsers()
	if errGettingList != nil {
		app.ErrorLog.Fatal()
		return
	}
	app.render(w, files, &templateData{
		UserData: userList,
	})

}

func (app *Application) ExpenseDetails(w http.ResponseWriter, r *http.Request) {
	expenseId, errConvert := strconv.Atoi(r.FormValue("expenseId"))
	if errConvert != nil {
		app.ErrorLog.Println(errConvert)
		log.Println(errConvert)
		return
	}
	expenseDetails, errDetails := app.Expense.ListExpensedetails(expenseId)
	if errDetails != nil {
		app.ErrorLog.Println()
		log.Println("AllUsers(): ", errDetails)
		return
	}

	files := []string{
		"ui/html/expensedetails.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	app.render(w, files, &templateData{
		UserId:         app.Session.GetInt(r, "userId"),
		ExpenseDetails: expenseDetails,
	})
}
