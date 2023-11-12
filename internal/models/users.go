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

func (m *UserModel) UserExists(id int) (bool, error) {
	return false, nil
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
