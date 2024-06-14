package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
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
	user := r.FormValue("email")
	password := r.FormValue("password")

	id, errAuth := app.User.Autenticate(user, password)
	if errAuth != nil {
		app.Session.Put(r, "flash", "Email or Password is incorrect")
		app.ErrorLog.Println(errAuth)
		log.Println(errAuth)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	app.Session.Put(r, "userId", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	name := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if app.Validate(r, name, "name") || app.Validate(r, email, "email") || app.Validate(r, password, "password") {
		http.Redirect(w, r, "/adduser", http.StatusSeeOther)
		return
	}

	if !app.isValidEmail(r, email) {
		http.Redirect(w, r, "/adduser", http.StatusSeeOther)
		return
	}

	err := app.User.InsertUser(name, email, password)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		http.Redirect(w, r, "/adduser", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/adduser", http.StatusSeeOther)
}

func (app *Application) isValidEmail(r *http.Request, email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@(gmail|yahoo)\.com$`
	if regexp.MustCompile(regex).MatchString(email) {
		exists, err := app.User.CheckEmail(email)
		if err != nil {
			app.ErrorLog.Println(err)
			app.Session.Put(r, "flash", "An error occurred while checking the email")
			return false
		}
		if exists {
			app.Session.Put(r, "flash", "The email already exists")
			return false
		}
		app.Session.Put(r, "flash", "User Successfully created")
		return true
	}
	app.Session.Put(r, "flash", "The email is not valid")
	return false
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
		Flash: app.Session.PopString(r, "flash"),
	})
}

func (app *Application) MarkAsPaid(w http.ResponseWriter, r *http.Request) {

	expenseId := r.FormValue("expenseId")
	intexpenseId, _ := strconv.Atoi(expenseId)
	userId := app.Session.GetInt(r, "userId")
	bool,err := app.Expense.CheckIfPaid(userId, intexpenseId)
	if err != nil {
		app.ErrorLog.Println()
	}
	if bool{
		app.Session.Put(r, "flash", "You already Paid Biaaatch")
		log.Println("You already Paid Biaaatch")
		http.Redirect(w, r, fmt.Sprintf("/expense_details?expenseId=%d", intexpenseId), http.StatusSeeOther)
		return
		
	}
	err = app.Expense.Mark(userId, intexpenseId)
	if err != nil {
		app.ErrorLog.Println()
	}
	http.Redirect(w, r, fmt.Sprintf("/expense_details?expenseId=%d", intexpenseId), http.StatusSeeOther)

}
