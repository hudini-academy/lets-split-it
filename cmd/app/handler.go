package main

import (
	"log"
	"net/http"
	"strconv"
)

// Home page for the application.
func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/home.page.tmpl",
		"ui/html/base.layout.tmpl",
	}
	s, err := app.Expense.GetYourSplit(app.Session.GetInt(r, "userId"))
	involvedSplits, errInvolved := app.Expense.GetInvolvedSplits(app.Session.GetInt(r, "userId"))
	if errInvolved!= nil {
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
	app.render(w, files, nil)
}

// Login the user after authentication.
func (app *Application) LoginUser(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("username")
	password := r.FormValue("password")

	id, errAuth := app.User.Autenticate(user, password)
	if errAuth != nil {
		app.ErrorLog.Println(errAuth)
		log.Println(errAuth)
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
	app.render(w, files, nil)
}

// AddUser adds a new user to the database.
func (app *Application) AddUser(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	err := app.User.InsertUser(name, email, password)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		log.Println("AddUser(): ", err)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	if title == ""{
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
		Flash: app.Session.PopString(r, "flash"),
	})

}

func (app *Application) ExpenseDetails(w http.ResponseWriter, r *http.Request) {
	expenseId, errConvert := strconv.Atoi(r.FormValue("expenseId"))
	if errConvert!= nil {
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
        UserId:    app.Session.GetInt(r, "userId"),
        ExpenseDetails: expenseDetails,
    })
}
