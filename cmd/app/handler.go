package main

import (
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
	log.Println(app.Session.GetString(r, "flash"))
	// Render the page with user list and flash message
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
	})
}

// DeleteUser is to delete the user already exists.
func (app *Application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	value, errConvert := strconv.Atoi(r.FormValue("userId"))
	if errConvert != nil {
		log.Println(errConvert)
		return
	}
	successDeleted, err := app.User.Delete(value)
	if err != nil {
		app.ErrorLog.Println(err.Error())
		log.Println("DeleteUser(): ", err)
		return
	}

	// Set appropriate flash message based on deletion success or involvement in splits
	if successDeleted {
		app.Session.Put(r, "flash", "User deleted successfully")
	} else if !successDeleted && err == nil {
<<<<<<< HEAD
		app.Session.Put(r, "Flash", "User is involved in a pending split. Cannot delete the user.")
=======
		app.Session.Put(r, "flash", "User is involved in a pending split. Cannot delete the user.")
>>>>>>> 0ea81ce445b6242cd7980ccbd173b18e2f35be31
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
