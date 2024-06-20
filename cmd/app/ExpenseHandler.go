package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// AddSplit handles adding a new split/expense to the database.
func (app *Application) AddSplit(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"ui/html/split.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	amount := r.FormValue("amount")
	note := r.FormValue("note")
	title := r.FormValue("title")
	usersSelected := r.Form["user[]"]

	app.Validate(r, amount, "amount")
	app.Validate(r, title, "title")
	flash := app.Session.PopString(r, "flash")
	userList, errGettingList := app.User.GetAllUsers()
	if errGettingList != nil {
		app.ErrorLog.Fatal()
		return
	}

	checkedUsers := make(map[int]bool)
	for _, id := range usersSelected {
		userID, _ := strconv.Atoi(id)
		checkedUsers[userID] = true
	}

	if app.Validate(r, amount, "amount") || app.Validate(r, title, "title") {
		app.render(w, files, &templateData{
			Flash:       flash,
			Amount:      amount,
			Title:       title,
			Description: note,
			UserData:    userList,
			CheckedUsers: checkedUsers,

		})
		return
	}
	// Parse form values
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusInternalServerError)
		return
	}

	// Check if users are selected for the split
	if len(usersSelected) == 0 {
		app.render(w, files, &templateData{
			Flash:        flash,
			Amount:       amount,
			Title:        title,
			Description:  note,
			UserData:     userList,
		})

		app.Session.Put(r, "flash", "No participants selected!")
		return
	}

	// Parse amount, note, and title from form values

	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		app.render(w, files, &templateData{
			Flash:       flash,
			Title:       title,
			Description: note,
			UserData:    userList,
			SelectedUsers: usersSelected,
			CheckedUsers: checkedUsers,

		})

		app.Session.Put(r, "flash", "Invalid amount!")
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

	app.render(w, files, &templateData{
		UserData:      userList,
		Flash:         app.Session.PopString(r, "flash"),
		TitleUserName: app.Session.GetString(r, "userName"),
	})
}

// ExpenseDetails displays details of an individual expense.
func (app *Application) ExpenseDetails(w http.ResponseWriter, r *http.Request) {
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

	files := []string{
		"ui/html/expensedetails.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	app.render(w, files, &templateData{
		UserId:         app.Session.GetInt(r, "userId"),
		ExpenseDetails: expenseDetails,
		Flash:          app.Session.PopString(r, "flash"),
		TitleUserName:  app.Session.GetString(r, "userName"),
	})
}

// MarkAsPaid handles marking an expense as paid by the user.
func (app *Application) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	expenseId := r.FormValue("expenseId")
	intexpenseId, _ := strconv.Atoi(expenseId)

	userId := app.Session.GetInt(r, "userId")

	// Check if expense is already paid by the user
	bool, err := app.Expense.CheckIfPaid(userId, intexpenseId)
	if err != nil {
		app.ErrorLog.Println(err)
	}

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

	http.Redirect(w, r, fmt.Sprintf("/expense_details?expenseId=%d", intexpenseId), http.StatusSeeOther)
}

// DeleteUser deletes a user from the database.
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

	if successDeleted {
		app.Session.Put(r, "flash", "User deleted successfully")
	} else if !successDeleted && err == nil {
		app.Session.Put(r, "flash", "User is involved in a pending split. Cannot delete the user.")
	}

	http.Redirect(w, r, "/allusers", http.StatusSeeOther)
}

// Cancelexpense cancels an expense from the database.
func (app *Application) Cancelexpense(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Allsplits retrieves and renders all splits that the user is involved in.
func (app *Application) Allsplits(w http.ResponseWriter, r *http.Request) {
	userId := app.Session.GetInt(r, "userId")

	files := []string{
		"ui/html/splitList.page.tmpl",
		"ui/html/base.layout.tmpl",
	}

	// Retrieve split transactions for the user.
	splitList, errFetchingSplitList := app.Expense.SplitList(userId)
	if errFetchingSplitList != nil {
		app.ErrorLog.Println(errFetchingSplitList.Error())
		log.Println("Allsplits(): ", errFetchingSplitList)
		return
	}

	app.render(w, files, &templateData{
		SplitTransaction: splitList,
		TitleUserName:    app.Session.GetString(r, "userName"),
	})
}
