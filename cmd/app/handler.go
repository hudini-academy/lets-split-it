package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Home renders the home page of the application.
func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	// Define template files to render
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

	// Render the page with retrieved data
	app.render(w, files, &templateData{
		UserId:        app.Session.GetInt(r, "userId"),
		TitleUserName: app.Session.GetString(r, "userName"),
		YourSplit:     s,
		Involved:      involvedSplits,
	})
}

// Login renders the login page.
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	// Define template files to render
	files := []string{
		"ui/html/login.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Render the page with flash message
	app.render(w, files, &templateData{
		Flash: app.Session.PopString(r, "flash"),
	})
}

// LoginUser handles user authentication.
func (app *Application) LoginUser(w http.ResponseWriter, r *http.Request) {
	// Define template files to render
	files := []string{
		"ui/html/login.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Retrieve email and password from form values
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
	// Define template files to render
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
	// Define template files to render
	files := []string{
		"ui/html/adduser.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Retrieve form values
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validate form fields
	app.Validate(r, username, "name")
	app.Validate(r, email, "email")
	app.Validate(r, password, "password")
	flash := app.Session.PopString(r, "flash")

	// If there are validation errors, render the template with errors
	if app.Validate(r, username, "name") || app.Validate(r, email, "email") || app.Validate(r, password, "password") {
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

	// Redirect to /adduser on success (to clear form)
	http.Redirect(w, r, "/adduser", http.StatusSeeOther)
}

// Validate performs field validation based on field type.
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

// Logout handles user logout by removing session data and redirecting to login page.
func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	app.Session.Remove(r, "userId")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// AllUsers retrieves and renders a list of all users.
func (app *Application) AllUsers(w http.ResponseWriter, r *http.Request) {
	// Retrieve list of all users
	userlist, err := app.User.ListUsers()
	if err != nil {
		app.ErrorLog.Println(err.Error())
		log.Println("AllUsers(): ", err)
		return
	}

	// Define template files to render
	files := []string{
		"ui/html/allusers.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Render the page with user list and flash message
	app.render(w, files, &templateData{
		UserList:      userlist,
		Flash:         app.Session.PopString(r, "flash"),
		TitleUserName: app.Session.GetString(r, "userName"),
	})
}

// AddSplit handles adding a new split/expense to the database.
func (app *Application) AddSplit(w http.ResponseWriter, r *http.Request) {
	// Parse form values
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	// Check if users are selected for the split
	usersSelected := r.Form["user[]"]
	if len(usersSelected) == 0 {
		app.Session.Put(r, "flash", "No participants selected!")
		http.Redirect(w, r, "/submit_expense", http.StatusSeeOther)
		return
	}

	// Parse amount, note, and title from form values
	amount := r.FormValue("amount")
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		app.Session.Put(r, "flash", "Invalid amount!")
		http.Redirect(w, r, "/submit_expense", http.StatusSeeOther)
		return
	}
	note := r.FormValue("note")
	title := r.FormValue("title")

	// Validate title field
	if title == "" {
		app.Session.Put(r, "flash", "Title Required!")
		http.Redirect(w, r, "/submit_expense", http.StatusSeeOther)
		return
	}

	// Insert expense into the database
	result, err := app.Expense.Insert(note, amountFloat, app.Session.GetInt(r, "userId"), title)
	if err != nil {
		log.Println(err)
		app.ErrorLog.Fatal()
		return
	}

	// Retrieve last inserted expense ID
	expenseId, err := result.LastInsertId()
	if err != nil {
		app.ErrorLog.Fatal()
	}

	// Insert splits associated with the expense
	app.Expense.Insert2Split(expenseId, amountFloat, usersSelected, app.Session.GetInt(r, "userId"))

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GetAddSplitForm retrieves and renders the form to add a new split/expense.
func (app *Application) GetAddSplitForm(w http.ResponseWriter, r *http.Request) {
	// Define template files to render
	files := []string{
		"ui/html/split.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Retrieve all users for the split form
	userList, errGettingList := app.User.GetAllUsers()
	if errGettingList != nil {
		app.ErrorLog.Fatal()
		return
	}

	// Render the page with user data, flash message, and username
	app.render(w, files, &templateData{
		UserData:      userList,
		Flash:         app.Session.PopString(r, "flash"),
		TitleUserName: app.Session.GetString(r, "userName"),
	})
}

// ExpenseDetails displays details of an individual expense.
func (app *Application) ExpenseDetails(w http.ResponseWriter, r *http.Request) {
	// Convert expenseId from string to integer
	expenseId, errConvert := strconv.Atoi(r.FormValue("expenseId"))
	if errConvert != nil {
		app.ErrorLog.Println(errConvert)
		log.Println(errConvert)
		return
	}

	// Retrieve expense details
	expenseDetails, errDetails := app.Expense.ListExpensedetails(expenseId, app.Session.GetInt(r, "userId"))
	if errDetails != nil {
		app.ErrorLog.Println(errDetails)
		log.Println("AllUsers(): ", errDetails)
		return
	}

	// Define template files to render
	files := []string{
		"ui/html/expensedetails.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Render the page with user ID, expense details, flash message, and username
	app.render(w, files, &templateData{
		UserId:         app.Session.GetInt(r, "userId"),
		ExpenseDetails: expenseDetails,
		Flash:          app.Session.PopString(r, "flash"),
		TitleUserName:  app.Session.GetString(r, "userName"),
	})
}

// MarkAsPaid handles marking an expense as paid by the user.
func (app *Application) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	// Retrieve expense ID from form value
	expenseId := r.FormValue("expenseId")
	intexpenseId, _ := strconv.Atoi(expenseId)

	// Retrieve user ID from session
	userId := app.Session.GetInt(r, "userId")

	// Check if expense is already paid by the user
	bool, err := app.Expense.CheckIfPaid(userId, intexpenseId)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	// Redirect with flash message if already paid
	if bool {
		app.Session.Put(r, "flash", "You already Paid!")
		http.Redirect(w, r, fmt.Sprintf("/expense_details?expenseId=%d", intexpenseId), http.StatusSeeOther)
		return
	}

	// Mark expense as paid
	err = app.Expense.Mark(userId, intexpenseId)
	if err != nil {
		app.ErrorLog.Println(err)
	}

	// Redirect to expense details page
	http.Redirect(w, r, fmt.Sprintf("/expense_details?expenseId=%d", intexpenseId), http.StatusSeeOther)
}

// DeleteUser deletes a user from the database.
func (app *Application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Retrieve user ID from form value
	value, errConvert := strconv.Atoi(r.FormValue("userId"))
	if errConvert != nil {
		log.Println(errConvert)
		return
	}

	// Attempt to delete user
	successDeleted, err := app.User.Delete(value)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		log.Println("DeleteUser(): ", err)
		app.Session.Put(r, "Flash", "Testing.")
		return
	}

	// Set appropriate flash message based on deletion success or involvement in splits
	if successDeleted {
		app.Session.Put(r, "Flash", "User deleted successfully")
	} else if !successDeleted && err == nil {
		app.Session.Put(r, "Flash", "User is involved in a pending split. Cannot delete the user.")
		log.Println("User is involved in a pending split. Cannot delete the user.")
	}

	// Redirect to all users page
	http.Redirect(w, r, "/allusers", http.StatusSeeOther)
}

// Cancelexpense cancels an expense from the database.
func (app *Application) Cancelexpense(w http.ResponseWriter, r *http.Request) {
	// Retrieve expense ID from form value
	value, errConvert := strconv.Atoi(r.FormValue("expenseId"))
	if errConvert != nil {
		log.Println(errConvert)
		return
	}

	// Cancel the expense
	err := app.Expense.Cancelupdate(value)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		log.Println("Cancelexpense(): ", err)
		return
	}

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Allsplits retrieves and renders all splits that the user is involved in.
func (app *Application) Allsplits(w http.ResponseWriter, r *http.Request) {
	// Retrieve user ID from session
	userId := app.Session.GetInt(r, "userId")

	// Define template files to render
	files := []string{
		"ui/html/splitList.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Retrieve split transactions for the user
	splitList, errFetchingSplitList := app.Expense.SplitList(userId)
	if errFetchingSplitList != nil {
		app.ErrorLog.Println(errFetchingSplitList.Error())
		log.Println("Allsplits(): ", errFetchingSplitList)
		return
	}

	// Render the page with split transactions, username
	app.render(w, files, &templateData{
		SplitTransaction: splitList,
		TitleUserName:    app.Session.GetString(r, "userName"),
	})
}
