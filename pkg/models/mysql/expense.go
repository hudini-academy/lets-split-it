package mysql

import (
	"database/sql"
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
