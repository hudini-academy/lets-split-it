package mysql

import (
	"database/sql"
	"expense/pkg/models"
	"log"
	"math"
	"strconv"
	"time"
)

type SplitModel struct {
	DB *sql.DB
}

func (m *SplitModel) Insert(note string, amount float64, userId int, title string) (sql.Result, error) {

	stmt := `INSERT INTO expense (note, amount,userId,date, title)
                VALUES(?,?,?,utc_timestamp(), ?)`

	result, err := m.DB.Exec(stmt, note, amount, userId, title)
	if err != nil {
		return nil, err
	}
	return result, nil

}

func (m *SplitModel) Insert2Split(ExpenseId int64, amount float64, userId []string, currentUserId int) error {

	splitAmount := amount / float64(len(userId))
	roundedSplitAmount := math.Round(splitAmount*100) / 100
	for _, user := range userId {
		userIdInt, _ := strconv.Atoi(user)
		// Insert split information into the split table
		if currentUserId == userIdInt {
			_, err := m.DB.Exec("INSERT INTO split (expenseId, userId, amount, datePaid) VALUES (?, ?, ?, UTC_TIMESTAMP())", ExpenseId, userIdInt, roundedSplitAmount)
			if err != nil {
				log.Println(err)
			}
		} else {
			_, err := m.DB.Exec("INSERT INTO split (expenseId, userId, amount) VALUES (?, ?, ?)", ExpenseId, userIdInt, roundedSplitAmount)
			if err != nil {
				log.Println(err)
			}
		}

	}
	return nil
}

func (m *SplitModel) GetYourSplit(userId int) ([]*models.Expense, error) {
	stmt := ` SELECT * FROM expense WHERE userId = ? `

	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sliceYourSplit := []*models.Expense{}
	for rows.Next() {
		s := &models.Expense{}
		err = rows.Scan(&s.ExpenseId, &s.UserId, &s.Note, &s.Amount, &s.Title, &s.Date, &s.Status)
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

// GetInvolvedSplits fetches the list of splits where the user have to pay.
func (m *SplitModel) GetInvolvedSplits(userId int) ([]*models.Expense, error) {
	var expenseDetails []*models.Expense
	stmt := `SELECT expenseId FROM split WHERE userId = ? AND datePaid IS NULL`
	rows, err := m.DB.Query(stmt, userId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		stmt2 := `SELECT * FROM expense WHERE expenseId =?`
		rows2 := m.DB.QueryRow(stmt2, id)
		expense := &models.Expense{}
		err = rows2.Scan(&expense.ExpenseId, &expense.UserId, &expense.Note, &expense.Amount, &expense.Title, &expense.Date, &expense.Status)
		if err != nil {
			return nil, err
		}
		expenseDetails = append(expenseDetails, expense)
	}
	return expenseDetails, nil
}

// ListExpensedetails returns the details of that specified expense.
func (m *SplitModel) ListExpensedetails(expenseId int) (*models.ExpenseDetails, error) {
	var totalAmount float64
	var note string
	var date time.Time
	var name string

	stmt := `SELECT amount, note, date, name FROM expense, user WHERE expenseId = ? AND expense.userId = user.userId`
	rows := m.DB.QueryRow(stmt, expenseId)

	rows.Scan(&totalAmount, &note, &date, &name)

	stmt2 := `SELECT amount, datePaid, name, split.userId, expenseId from split, user WHERE split.userId = user.userId AND split.expenseId = ?`
	rows2, err := m.DB.Query(stmt2, expenseId)
	if err != nil {
		return nil, err
	}
	var splitDetails []*models.Split
	for rows2.Next() {
		s := &models.Split{}
		err = rows2.Scan(&s.Amount, &s.DatePaid, &s.Name, &s.UserId, &s.ExpenseId)
		if err != nil {
			return nil, err
		}
		splitDetails = append(splitDetails, s)
	}
	var expenseDetails *models.ExpenseDetails
	expenseDetails = &models.ExpenseDetails{
		Amount:       totalAmount,
		Note:         note,
		Date:         date,
		CreatedName:  name,
		SplitDetails: splitDetails,
	}
	return expenseDetails, nil
}
