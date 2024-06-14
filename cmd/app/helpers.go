package main

import (
	"database/sql"
	"expense/pkg/models"
	"log"
	"net/http"
	"os"
	"regexp"
	"text/template"
)

type templateData struct {
	Flash          string
	Error          map[string]string
	Email          string
	Username       string
	UserId         int
	UserList       []*models.User
	UserData       []*models.User
	YourSplit      []*models.Expense
	Involved       []*models.Expense
	ExpenseDetails *models.ExpenseDetails
}

// LogFiles opens the log files and return them.
func LogFiles() (*log.Logger, *log.Logger, error) {
	// Open file to store information messages.
	infoFile, errOpenInfo := os.OpenFile("./tmp/info.log", os.O_RDWR|os.O_CREATE, 0666)
	if errOpenInfo != nil {
		return nil, nil, errOpenInfo
	}

	// Open file to store error messages.
	errorFile, errOpenError := os.OpenFile("./tmp/error.log", os.O_RDWR|os.O_CREATE, 0666)
	if errOpenError != nil {
		return nil, nil, errOpenError
	}

	return log.New(infoFile, "INFO: ", log.Ldate|log.Ltime), log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile), nil
}

// OpenDb creates a new database connection and test it.
func openDB(dsn string) (*sql.DB, error) {
	db, errOpenDb := sql.Open("mysql", dsn)
	if errOpenDb != nil {
		return nil, errOpenDb
	}
	return db, nil
}

// render function displays the page.
func (app *Application) render(w http.ResponseWriter, files []string, td *templateData) {
	ts, errParsingFiles := template.ParseFiles(files...)

	//  checking for any error
	if errParsingFiles != nil {
		app.ErrorLog.Println(errParsingFiles)
		log.Println(errParsingFiles)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Executing the template and checking for any error
	errExecutingFiles := ts.Execute(w, td)
	if errExecutingFiles != nil {
		app.ErrorLog.Println(errExecutingFiles)
		log.Println(errExecutingFiles)
		http.Error(w, "Internal Server error", 500)
		return
	}
}

// AutenticatedUser checks whether the user is authenticated or not.
// Redirects to login page if the user is not authenticated.
func (app *Application) AuthenticatedUser(r *http.Request) bool {
	loggedId := app.Session.GetInt(r, "userId")
	return loggedId != 0
}

func (app *Application) isValidEmail(r *http.Request, email string) bool {
	regex := `^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`
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

func (app *Application) isValidUser(r *http.Request, name string) bool {
	exists, err := app.User.CheckUser(name)
	if err != nil {
		app.ErrorLog.Println(err)
		app.Session.Put(r, "flash", "An error occurred while checking the email")
		return false
	}

	if exists {
		app.Session.Put(r, "flash", "The name already exists")
		return false
	}
	return true
}
