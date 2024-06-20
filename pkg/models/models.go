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

// User represents a user in the system.
type User struct {
	UserID         int
	Name           string
	Email          string
	HashedPassword []byte
	Role           int
}

// Expense represents an expense in the system.
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

//	Split represents a single split in the system.
type Split struct {
	ExpenseId int
	UserId    int
	Amount    float64
	DatePaid  sql.NullTime
	Name      string
}

// ExpenseDetails represents details of an expense in the system.
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
	Status             int
}
