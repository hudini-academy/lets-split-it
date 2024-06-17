package mysql

import (
	"database/sql"
	"expense/pkg/models"
	"log"
	"math"
	"strconv"
	"time"
)

// SplitModel handles database operations related to expenses and splits.
type SplitModel struct {
	DB *sql.DB
}

// Insert inserts a new expense record into the database.

// Parameters:
// note - description for the new expense record.
// amount - Total amount spent in the expense.
// userId - ID of the user who spent the expense.
// title - Title of the expense.

// Returns: The sql result and the error.
func (m *SplitModel) Insert(note string, amount float64, userId int, title string) (sql.Result, error) {
	stmt := `INSERT INTO expense (note, amount, userId, date, title)
				VALUES(?, ?, ?, utc_timestamp(), ?)`

	result, err := m.DB.Exec(stmt, note, amount, userId, title)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Insert2Split inserts split details into the database for an expense.

// Parameters:
// expenseId - ID of the expense for which split details are being inserted.
// amount - Total amount to be split among users.
// userId - List of user IDs to split the expense.
// currentUserId - ID of the user initiating the split.

// Returns: Error, if any.
func (m *SplitModel) Insert2Split(expenseId int64, amount float64, userId []string, currentUserId int) error {
	splitAmount := amount / float64(len(userId))
	roundedSplitAmount := math.Round(splitAmount*100) / 100

	for _, user := range userId {
		userIdInt, _ := strconv.Atoi(user)

		// Insert split information into the split table
		if currentUserId == userIdInt {
			_, err := m.DB.Exec("INSERT INTO split (expenseId, userId, amount, datePaid) VALUES (?, ?, ?, UTC_TIMESTAMP())", expenseId, userIdInt, roundedSplitAmount)
			if err != nil {
				log.Println(err)
			}
			_, err = m.DB.Exec("UPDATE expense SET status = 1 WHERE expenseId = ?", expenseId)
			if err != nil {
				log.Println(err)
			}
		} else {
			_, err := m.DB.Exec("INSERT INTO split (expenseId, userId, amount) VALUES (?, ?, ?)", expenseId, userIdInt, roundedSplitAmount)
			if err != nil {
				log.Println(err)
			}
		}
	}
	return nil
}

// GetYourSplit fetches expenses created by the specified user.

// Parameters:
// userId - ID of the user whose expenses are to be fetched.

// Returns: A slice of Expense pointers and an error, if any.
func (m *SplitModel) GetYourSplit(userId int) ([]*models.Expense, error) {
	stmt := `SELECT e.*, u.name FROM expense e, user u WHERE e.userId = ? AND u.userId = e.userId ORDER BY e.expenseId DESC`

	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sliceYourSplit := []*models.Expense{}
	for rows.Next() {
		s := &models.Expense{}
		err = rows.Scan(&s.ExpenseId, &s.UserId, &s.Note, &s.Amount, &s.Title, &s.Date, &s.Status, &s.CreatedUserName)
		if err != nil {
			return nil, err
		}
		sliceYourSplit = append(sliceYourSplit, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sliceYourSplit, nil
}

// GetInvolvedSplits fetches expenses where the user is involved.

// Parameters:
// userId - ID of the user involved in the expenses.

// Returns: A slice of Expense pointers and an error, if any.
func (m *SplitModel) GetInvolvedSplits(userId int) ([]*models.Expense, error) {
	var expenseDetails []*models.Expense
	stmt := `SELECT expenseId FROM split WHERE userId = ? AND datePaid IS NULL ORDER BY expenseId DESC`
	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		stmt2 := `SELECT e.*, u.name FROM expense e, user u WHERE expenseId = ? AND u.userId = e.userId`
		rows2 := m.DB.QueryRow(stmt2, id)
		expense := &models.Expense{}
		err = rows2.Scan(&expense.ExpenseId, &expense.UserId, &expense.Note, &expense.Amount, &expense.Title, &expense.Date, &expense.Status, &expense.CreatedUserName)
		if err != nil {
			return nil, err
		}
		expenseDetails = append(expenseDetails, expense)
	}
	return expenseDetails, nil
}

// ListExpensedetails returns details of a specific expense.

// Parameters:
// expenseId - ID of the expense whose details are to be fetched.
// userId - ID of the user fetching the expense details.

// Returns: ExpenseDetails pointer and an error, if any.
func (m *SplitModel) ListExpensedetails(expenseId, userId int) (*models.ExpenseDetails, error) {
	var totalAmount float64
	var note string
	var date time.Time
	var name string
	var title string
	var expenseUserId int
	var status int
	var paid int // 0 or 1 based on whether the user paid the expense

	stmt := `SELECT amount, note, date, name, title, expense.userId, status FROM expense, user WHERE expenseId = ? AND expense.userId = user.userId`
	rows := m.DB.QueryRow(stmt, expenseId)

	rows.Scan(&totalAmount, &note, &date, &name, &title, &expenseUserId, &status)

	if userId == expenseUserId {
		paid = 1
	} else {
		paid = 0
	}

	stmt2 := `SELECT amount, datePaid, name, split.userId, expenseId from split, user WHERE split.userId = user.userId AND split.expenseId = ?`
	rows2, err := m.DB.Query(stmt2, expenseId)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	var splitDetails []*models.Split
	var outstandingBalance float64

	for rows2.Next() {
		s := &models.Split{}
		err = rows2.Scan(&s.Amount, &s.DatePaid, &s.Name, &s.UserId, &s.ExpenseId)
		if err != nil {
			return nil, err
		}

		if s.UserId == userId && s.DatePaid.Valid {
			paid = 1
		}

		if !s.DatePaid.Valid {
			outstandingBalance += s.Amount
		}

		splitDetails = append(splitDetails, s)
	}

	expenseDetails := &models.ExpenseDetails{
		ExpenseId:          expenseId,
		Amount:             totalAmount,
		Note:               note,
		Date:               date,
		CreatedName:        name,
		Title:              title,
		SplitDetails:       splitDetails,
		OutstandingBalance: outstandingBalance,
		Paid:               paid,
		Status:             status,
	}
	return expenseDetails, nil
}

// Cancelupdate updates the status of an expense to cancelled.

// Parameters:
// expenseId - ID of the expense to be cancelled.

// Returns: Error, if any.
func (m *SplitModel) Cancelupdate(expenseId int) error {
	stmt := `UPDATE expense SET status = 2 WHERE expenseId = ?`

	_, err := m.DB.Exec(stmt, expenseId)
	if err != nil {
		return err
	}
	return nil
}

// Mark updates the payment status of an expense and its splits.

// Parameters:
// userId - ID of the user marking the expense as paid.
// expenseId - ID of the expense being marked.

// Returns: Error, if any.
func (m *SplitModel) Mark(userId int, expenseId int) error {
	// Check the count of non-null values in datepaid
	stmtCheckAllNull := `SELECT COUNT(*) FROM split WHERE expenseId = ? AND datepaid IS NOT NULL`
	var nonNullCount int
	err := m.DB.QueryRow(stmtCheckAllNull, expenseId).Scan(&nonNullCount)
	if err != nil {
		return err
	}

	// If all datepaid fields are NULL, set the status to 1
	if nonNullCount == 0 {
		stmtUpdateStatus := `UPDATE expense SET status = 1 WHERE expenseId = ?`
		_, err = m.DB.Exec(stmtUpdateStatus, expenseId)
		if err != nil {
			return err
		}
	}

	// Update the datepaid of the corresponding user
	stmtUpdate := `UPDATE split SET datepaid = UTC_TIMESTAMP() WHERE userId = ? AND expenseId = ?`
	_, err = m.DB.Exec(stmtUpdate, userId, expenseId)
	if err != nil {
		return err
	}

	// Check if all users in the expense have non-null datepaid
	stmtCheck := `SELECT COUNT(*) FROM split WHERE expenseId = ? AND datepaid IS NULL`
	var count int
	err = m.DB.QueryRow(stmtCheck, expenseId).Scan(&count)
	if err != nil {
		return err
	}

	// If all users have a non-null datepaid, update the expense status to 2
	if count == 0 {
		stmtUpdateStatus := `UPDATE expense SET status = 2 WHERE expenseId = ?`
		_, err = m.DB.Exec(stmtUpdateStatus, expenseId)
		if err != nil {
			return err
		}
	}

	// Check whether all users paid the split.
	// If paid, change the status of the expense to completed.
	stmtCheckPaid := `SELECT COUNT(*) FROM split WHERE expenseId = ? AND datepaid IS NULL`
	var countPaid int
	err = m.DB.QueryRow(stmtCheckPaid, expenseId).Scan(&countPaid)
	if err != nil {
		return err
	}
	if countPaid == 0 {
		stmtUpdateStatus := `UPDATE expense SET status = 3 WHERE expenseId = ?`
		_, err = m.DB.Exec(stmtUpdateStatus, expenseId)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckIfPaid checks if a user has paid a specific expense.

// Parameters:
// userId - ID of the user to check.
// expenseId - ID of the expense to check.

// Returns: Boolean indicating whether the user has paid and an error, if any.
func (m *SplitModel) CheckIfPaid(userId int, expenseId int) (bool, error) {
	var datePaid sql.NullTime
	stmt := `SELECT datepaid FROM split WHERE userId = ? AND expenseId = ?`

	err := m.DB.QueryRow(stmt, userId, expenseId).Scan(&datePaid)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // no record found
		}
		return false, err
	}

	// it means datepaid is not null
	return datePaid.Valid, nil
}

// SplitList returns the list of expenses where the user is involved.

// Parameters:
// userId - ID of the user whose involved expenses are to be fetched.

// Returns: A slice of ExpenseDetails pointers and an error, if any.
func (m *SplitModel) SplitList(userId int) ([]*models.ExpenseDetails, error) {
	var expenseDetails []*models.ExpenseDetails
	stmt := `SELECT distinct expenseId FROM split WHERE userId = ? ORDER BY expenseId DESC`
	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		stmt2 := `SELECT e.expenseId, e.note, e.amount, e.title, e.date, e.status, u.name FROM expense e, user u WHERE expenseId =? AND u.userId = e.userId`
		rows2 := m.DB.QueryRow(stmt2, id)
		expense := &models.ExpenseDetails{}
		err = rows2.Scan(&expense.ExpenseId, &expense.Note, &expense.Amount, &expense.Title, &expense.Date, &expense.Status, &expense.CreatedName)
		if err != nil {
			return nil, err
		}
		expenseDetails = append(expenseDetails, expense)
	}
	return expenseDetails, nil
}
