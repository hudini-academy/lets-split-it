package main

import (
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
		Flash:    app.Session.PopString(r, "flash"),
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
		app.Session.Put(r, "flash", "Invalid amount !")
		http.Redirect(w, r, "/submit_expense", http.StatusSeeOther)
		return
	}
	note := r.FormValue("note")
	title := r.FormValue("title")

	if title == "" {
		app.Session.Put(r, "flash", "Title Required !")
		http.Redirect(w, r, "/submit_expense", http.StatusSeeOther)
		return
	}

	result, err := app.Expense.Insert(note, amountFloat, app.Session.GetInt(r, "userId"), title)
	http.Redirect(w, r, "/submit_expense", http.StatusSeeOther)
	if err != nil {
		log.Println(err)
		app.ErrorLog.Fatal()
		return
	}
	app.Session.Put(r, "flash", "Task successfully created !")

	usersSelected := r.Form["user[]"]

	expenseId, err := result.LastInsertId()
	if err != nil {
		app.ErrorLog.Fatal()
	}
	app.Expense.Insert2Split(expenseId, amountFloat, usersSelected, app.Session.GetInt(r, "userId"))
	http.Redirect(w, r, "/", http.StatusSeeOther)

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
		Flash:    app.Session.PopString(r, "flash"),
	})

}

// ExpenseDetails display the details of indiviual expense
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
		Flash:          app.Session.PopString(r, "flash"),
	})
}

func (app *Application) MarkAsPaid(w http.ResponseWriter, r *http.Request) {

	expenseId := r.FormValue("expenseId")
	intexpenseId, _ := strconv.Atoi(expenseId)
	log.Println(intexpenseId)
	userId := app.Session.GetInt(r, "userId")
	bool, err := app.Expense.CheckIfPaid(userId, intexpenseId)
	if err != nil {
		app.ErrorLog.Println()
	}
	if bool {
		app.Session.Put(r, "flash", "You already Paid !")
		http.Redirect(w, r, fmt.Sprintf("/expense_details?expenseId=%d", intexpenseId), http.StatusSeeOther)
		return

	}
	err = app.Expense.Mark(userId, intexpenseId)
	if err != nil {
		app.ErrorLog.Println()
	}
	http.Redirect(w, r, fmt.Sprintf("/expense_details?expenseId=%d", intexpenseId), http.StatusSeeOther)

}

// DeleteUser is to delete the user already exists.
func (app *Application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	value, errConvert := strconv.Atoi(r.FormValue("userId"))
	if errConvert != nil {
		log.Println(errConvert)
		return
	}
	successDeleted, err := app.User.Delete(value)
	if successDeleted {
		app.Session.Put(r, "Flash", "User deleted successfully")
	} else if !successDeleted && err == nil {
		app.Session.Put(r, "Flash", "User is involved in a pending split. Cannot delete the user.")
	}
	if err != nil {
		app.ErrorLog.Println(err.Error())
		app.Session.Put(r, "Flash", "Testing.")
		log.Println("DeleteUser(): ", err)
		return
	}
	app.Session.Put(r, "Flash", "Testing.")
	// Redirecting to the all users page by using http.Redirect
	http.Redirect(w, r, "/allusers", http.StatusSeeOther)
}

// Cancelexpense is to cnacel the expense that created
func (app *Application) Cancelexpense(w http.ResponseWriter, r *http.Request) {
	value, errConvert := strconv.Atoi(r.FormValue("expenseId"))
	if errConvert != nil {
		log.Println(errConvert)
		return
	}

	err := app.Expense.Cancelupdate(value)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		log.Println("Cancelexpense(): ", err)
		return
	}

	// redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
