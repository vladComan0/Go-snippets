package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

const (
	COST          = 12   // 2^12 bcrypt iterations used to generate the password hash (4-31)
	ERR_DUP_ENTRY = 1062 // MySQL Error number for duplicate entries
	CONSTRAINT    = "user_uc_email"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (*User, error)
	UpdatePassword(id int, currentPassword, newPassword string) error
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), COST)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (name, email, hashed_password, created)
	VALUES (?, ?, ?, UTC_TIMESTAMP())`
	_, err = m.DB.Exec(stmt, name, email, hashedPassword)
	if err != nil {
		// Validate also the duplicate email error
		if validateEmailError := validateDuplicateEmail(err); validateEmailError != nil {
			return validateEmailError
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var (
		id             int
		hashedPassword []byte
	)
	stmt := "SELECT id, hashed_password FROM users WHERE email = ?"
	if err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrInvalidCredentials
		default:
			return 0, err
		}
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return 0, ErrInvalidCredentials
		default:
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = ?)"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)

	return exists, err
}

func (m *UserModel) Get(id int) (*User, error) {
	user := &User{}

	stmt := "SELECT id, name, email, created FROM users WHERE id = ?"
	if err := m.DB.QueryRow(stmt, id).Scan(&user.ID, &user.Name, &user.Email, &user.Created); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoRecord
		default:
			return nil, err
		}
	}

	return user, nil
}

func (m *UserModel) UpdatePassword(id int, currentPassword, newPassword string) error {
	var hashedCurrentPassword []byte

	stmt := "SELECT hashed_password FROM users WHERE id = ?"
	if err := m.DB.QueryRow(stmt, id).Scan(&hashedCurrentPassword); err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(hashedCurrentPassword, []byte(currentPassword)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return ErrInvalidCredentials
		default:
			return err
		}
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), COST)
	if err != nil {
		return err
	}

	stmt = "UPDATE users SET hashed_password = ? WHERE id = ?"
	if _, err := m.DB.Exec(stmt, hashedNewPassword, id); err != nil {
		return err
	}

	return nil
}

func validateDuplicateEmail(err error) error {
	var mySQLError *mysql.MySQLError
	if errors.As(err, &mySQLError) {
		if mySQLError.Number == ERR_DUP_ENTRY && strings.Contains(mySQLError.Message, CONSTRAINT) {
			return ErrDuplicateEmail
		}
	}
	return nil
}
