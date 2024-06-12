package main

import (
	"database/sql"
	"expense/pkg/models"
	"log"
	"net/http"
	"os"
	"text/template"
)

type templateData struct {
	Flash    string
	Error    map[string]string
	UserId   int
	UserList []*models.User
	UserData []*models.User
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
