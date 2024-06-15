package models

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
)

type User struct {
	UserID         int
	Name           string
	Email          string
	HashedPassword []byte
	Role           int
}

type Expense struct {
	ExpenseId       int
	UserId          int
	CreatedUserName string
	Note            string
	Amount          float64
	Title           string
	Date            time.Time
	Status          int
}

type Split struct {
	ExpenseId int
	UserId    int
	Amount    float64
	DatePaid  sql.NullTime
	Name      string
}

type ExpenseDetails struct {
	ExpenseId          int
	Amount             float64
	Date               time.Time
	Note               string
	CreatedName        string
	Title              string
	OutstandingBalance float64
	SplitDetails       []*Split
	Paid               int
}

func (e ExpenseDetails) formatDate(date time.Time) string {
	return date.Format("02 Jan 2006")
}
