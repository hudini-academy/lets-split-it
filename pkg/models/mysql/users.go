package mysql

import (
	"database/sql"
	"expense/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

// InsertUser creates a new user.
func (m *UserModel) InsertUser(name, email, password string) error {
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := ` insert into user (name , email , password) values (?,?,?)`
	_, Inserterr := m.DB.Exec(stmt, name, email, hashedpassword)
	if Inserterr != nil {
		return Inserterr
	}
	return nil
}

// Authenticate function checks if the user is in the datavase and returns the id of the user.
func (u *UserModel) Autenticate(username, password string) (int, error) {
	var id int
	var hashedPassword []byte
	row := u.DB.QueryRow(`SELECT userId, password FROM user WHERE email = ?`, username)
	err := row.Scan(&id, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	// Compare the provided password with the hashed password. If they match.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, models.ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}
	return id, nil
}

// ListUsers returns all the users in the database.
func (m *UserModel) ListUsers() ([]*models.User, error) {
	stmt := ` SELECT  userId, name ,email from user where userId > 1 `

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sliceUser := []*models.User{}
	for rows.Next() {
		s := &models.User{}
		err = rows.Scan(&s.UserID, &s.Name, &s.Email)
		if err != nil {
			return nil, err
		}
		sliceUser = append(sliceUser, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sliceUser, nil

}

func (m *UserModel) GetAllUsers() ([]*models.User, error) {
	stmt := `SELECT userId,name from user`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	UsersinDB := []*models.User{}

	for rows.Next() {
		// Create a pointer to a new zeroed Todos struct.
		s := &models.User{}

		err = rows.Scan(&s.UserID, &s.Name)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of todos.
		UsersinDB = append(UsersinDB, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return UsersinDB, nil
}