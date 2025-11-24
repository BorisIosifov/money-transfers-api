package model

import (
	"fmt"
	"log"
)

const (
	getUser = `
SELECT user_id, email, phone, type, external_user_id, telegram_chat_id, name, role, ctime
FROM users
WHERE email = $1 AND password = $2`

	checkExistingUser = `SELECT count(*) FROM users WHERE email = $1`

	createUser = `
INSERT INTO users (email, password, phone, type, external_user_id, telegram_chat_id, name, role)
VALUES (:email, :password, :phone, :type, :external_user_id, :telegram_chat_id, :name, :role)
RETURNING user_id`

	updateUserPassword = `UPDATE users SET password = $1 WHERE email = $2`
)

// swagger:model
type User struct {
	UserID            int    `db:"user_id"`
	Email             string `db:"email"`
	Password          string `db:"password" json:"-"`
	Phone             string `db:"phone"`
	Type              string `db:"type"`
	ExternalUserID    int    `db:"external_user_id"`
	TelegramChatID    int    `db:"telegram_chat_id"`
	Name              string `db:"name"`
	Role              string `db:"role"`
	CTime             string `db:"ctime"`
	PasswordUncrypted string `json:"-"`
}

func (user User) Create(tx TXWrapper) (User, error) {
	user.Password = fmt.Sprintf("%x", []byte(user.PasswordUncrypted))
	rows, err := tx.NamedQuery(createUser, user)
	if err != nil {
		return user, err
	}
	defer rows.Close()

	var LastInsertId int
	if rows.Next() {
		err = rows.Scan(&LastInsertId)
		if err != nil {
			return user, err
		}
	}
	log.Printf("LastInsertId: %d", LastInsertId)
	user.UserID = LastInsertId

	return user, err
}

func GetUserByEmailAndPassword(db DBWrapper, Email, Password string) (User, error) {
	var passwordEncrypted = fmt.Sprintf("%x", []byte(Password))
	var user User
	err := db.Get(&user, getUser, Email, passwordEncrypted)
	return user, err
}

func DoesUserExist(db DBWrapper, Email string) (bool, error) {
	var countUsers int
	err := db.Get(&countUsers, checkExistingUser, Email)
	return countUsers > 0, err
}

func (user User) UpdatePassword(tx TXWrapper) error {
	user.Password = fmt.Sprintf("%x", []byte(user.PasswordUncrypted))
	_, err := tx.Exec(updateUserPassword, user.Password, user.Email)
	return err
}
