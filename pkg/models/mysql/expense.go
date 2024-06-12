package mysql

import (
	"database/sql"
	"expense/pkg/models"
	"log"
	"math"
	"strconv"
)

type SplitModel struct {
	DB *sql.DB
}

func (m *SplitModel) Insert(note string, amount float64, userId int) (sql.Result, error) {

	stmt := `INSERT INTO expense (note, amount,userId,date)
				VALUES(?,?,?,utc_timestamp())`

	result, err := m.DB.Exec(stmt, note, amount, userId)
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

func (m *SplitModel) GetYourSplit(userId int)([]*models.Expense, error) {
	stmt := ` SELECT * FROM expense WHERE userId = ? `
	log.Println(userId)

	rows, err := m.DB.Query(stmt,userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sliceYourSplit := []*models.Expense{}
	for rows.Next() {
		s := &models.Expense{}
		err = rows.Scan(&s.ExpenseId,&s.Note,&s.Amount, &s.Date, &s.UserId,&s.Status)
		log.Println("inside scan")
		if err != nil {
			return nil, err
		}
		sliceYourSplit = append(sliceYourSplit, s)
		log.Println(sliceYourSplit)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sliceYourSplit, nil

}
