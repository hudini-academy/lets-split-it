package main

import (
	"database/sql"
	"expense/pkg/models"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"
)

// templateData holds data that is passed to templates for rendering HTML pages.
type templateData struct {
	Flash            string                   // Flash message to display to the user
	Error            map[string]string        // Map of errors with specific fields
	Email            string                   // Email address associated with the user - retaining email while login.
	Username         string                   // Username associated with the user - retains username while login.
	UserId           int                      // User ID associated with the logged-in user
	UserList         []*models.User           // List of users
	UserData         []*models.User           // Data related to a specific user
	YourSplit        []*models.Expense        // List of expenses created by the user
	Involved         []*models.Expense        // List of expenses where the user is involved
	ExpenseDetails   *models.ExpenseDetails   // Details of a specific expense
	SplitTransaction []*models.ExpenseDetails // Split history of the user.
	TitleUserName    string
	Title            string // Display title content in create split page
	Description      string // Display description in create split page
	Amount           string // Display Amount in create split page
	SelectedUsers    []string
	CheckedUsers     map[int]bool
	CurrentPassword  string
}

// LogFiles opens the log files for writing information and error messages.
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

	// Return loggers for information and error messages.
	return log.New(infoFile, "INFO: ", log.Ldate|log.Ltime), log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile), nil
}

// openDB creates a new database connection.
func openDB(dsn string) (*sql.DB, error) {
	db, errOpenDb := sql.Open("mysql", dsn)
	if errOpenDb != nil {
		return nil, errOpenDb
	}
	return db, nil
}

// render function renders the specified template files with the provided template data (td).
func (app *Application) render(w http.ResponseWriter, files []string, td *templateData) {

	ts, errParsingFiles := template.ParseFiles(files...)

	// Check for any error while parsing template files.
	if errParsingFiles != nil {
		app.ErrorLog.Println(errParsingFiles)
		log.Println(errParsingFiles)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Execute the parsed template and handle any execution error.
	errExecutingFiles := ts.Execute(w, td)
	if errExecutingFiles != nil {
		app.ErrorLog.Println(errExecutingFiles)
		log.Println(errExecutingFiles)
		http.Error(w, "Internal Server error", 500)
		return
	}

}

// AuthenticatedUser checks whether the user is authenticated based on the session data.
// It returns true if the user is authenticated, otherwise false.
func (app *Application) AuthenticatedUser(r *http.Request) bool {
	loggedId := app.Session.GetInt(r, "userId")
	return loggedId != 0
}

// isValidEmail checks if the provided email address is valid and not already in use.
// It returns true if the email is valid and available, otherwise false.
func (app *Application) isValidEmail(r *http.Request, email string) bool {
	regex := `^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`

	// Validate email format using regular expression.
	if regexp.MustCompile(regex).MatchString(email) {
		// Check if the email already exists in the database.
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
		// Indicate successful creation of the user.
		app.Session.Put(r, "flash", "User Successfully created")
		return true
	}
	// If email format is invalid, set flash message and return false.
	app.Session.Put(r, "flash", "The email is not valid")
	return false
}

// isValidUser checks if the provided username is valid and not already in use.
// It returns true if the username is valid and available, otherwise false.
func (app *Application) isValidUser(r *http.Request, name string) bool {
	// Check if the username already exists in the database.
	exists, err := app.User.CheckUser(name)
	if err != nil {
		app.ErrorLog.Println(err)
		app.Session.Put(r, "flash", "An error occurred while checking the email")
		return false
	}

	// If the username exists, set flash message and return false.
	if exists {
		app.Session.Put(r, "flash", "The name already exists")
		return false
	}
	return true
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
	case "title":
		if strings.TrimSpace(field) == "" {
			app.Session.Put(r, "flash", "The Title field is blank!")
			return true
		}
	case "amount":
		if strings.TrimSpace(field) == "" {
			app.Session.Put(r, "flash", "The Amount field is blank!")
			return true
		}
	}
	return false
}
