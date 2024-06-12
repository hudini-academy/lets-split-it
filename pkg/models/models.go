package models

import (
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
	ExpenseId int
	UserId    int
	Note      string
	Amount    float64
	Date      time.Time
}

type Split struct {
	ExpenseId int
	UserId    int
	Amount    float64
	DatePaid  time.Time
}
