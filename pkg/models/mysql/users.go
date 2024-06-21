package mysql

import (
	"database/sql"
	"expense/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// UserModel handles database operations related to the user.
type UserModel struct {
	DB *sql.DB
}

// InsertUser creates a new user.

// Parameters:
// name - the name of the user.
// email - the email of the user.
// password - the password of the user.

// Returns: Error, if any.
func (m *UserModel) InsertUser(name, email, password string) error {
	// Encrypt the password.
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

// CheckEmail checks if the email is already in the database.

// Parameters:
// email - The email to check.

// Returns: true if the email is in the database, false otherwise and an error, if any.
func (m *UserModel) CheckEmail(email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM user WHERE email = ?)"
	err := m.DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// CheckUser checks if the user is already in the database.

// Parameters:
// name - name of the user.

// Returns: true if the user is in the database, false otherwise and an error, if any.
func (m *UserModel) CheckUser(name string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM user WHERE name = ?)"
	err := m.DB.QueryRow(query, name).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Authenticate function checks if the user is in the database.

// Parameters:
// username - the username of the user.
// password - the password of the user.

// Returns: the id of the user, the name of the user and an error, if any.
func (u *UserModel) Autenticate(username, password string) (int, string, error) {
	var id int
	var name string
	var hashedPassword []byte
	row := u.DB.QueryRow(`SELECT userId, name, password FROM user WHERE email = ?`, username)
	err := row.Scan(&id, &name, &hashedPassword)
	if err == sql.ErrNoRows {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		return 0, "", err
	}

	// Compare the provided password with the hashed password. If they match.
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		return 0, "", err
	}
	return id, name, nil
}

// ListUsers fetches all the users in the database otherthan admin.

// Parameters: none

// Returns: a slice of users and an error, if any.
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

// GetAllUsers fetches all the users in the database.

// Parameters: none

// Returns: a slice of users and an error, if any.
func (m *UserModel) GetAllUsers() ([]*models.User, error) {
	stmt := `SELECT userId,name from user`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	UsersinDB := []*models.User{}

	for rows.Next() {
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

// Delete will remove the user from the database.
// The user will be removed only if the user is not involved in an active split.

// Parameters:
// id - the id of the user.

// Returns: true if the user is deleted, false otherwise and an error, if any.
func (m *UserModel) Delete(id int) (bool, error) {
	stmt := `DELETE user
				 FROM user
				 LEFT JOIN split ON user.userId = split.userId AND split.datePaid IS NULL
				 WHERE user.userId = ? AND split.userId IS NULL;`

	result, err := m.DB.Exec(stmt, id)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

// ChangePassword updates the password of the user.

// Parameters:
// userId - The id of the user.
// newPassword - The new password.

// Returns: true if the user is deleted, false otherwise and an error, if any.
func (m *UserModel) ChangePassword(userId int, currentPassword, newPassword string) (bool, error) {
	var passwordFromDb []byte
	stmt := `SELECT password FROM user WHERE userId = ?`
	row := m.DB.QueryRow(stmt, userId)
	row.Scan(&passwordFromDb)
	err := bcrypt.CompareHashAndPassword(passwordFromDb, []byte(currentPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, models.ErrInvalidCredentials
	} else if err != nil {
		return false, err
	}
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return false, err
	}
	stmt = `UPDATE user SET password = ?  WHERE userId = ?`

	_, err = m.DB.Exec(stmt, hashedpassword, userId)
	if err != nil {
		return false, err
	}

	return true, nil
}
