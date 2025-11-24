package model

import "time"

const (
	createEmailCode = `INSERT INTO email_codes (email, code, code_type) VALUES (:email, :code, :code_type)`

	getCodeByEmail = `SELECT code_id, email, code, ctime, code_type, attempts FROM email_codes WHERE email = $1 AND code_type = $2 ORDER BY ctime DESC LIMIT 1`

	increaseAttemptsInEmailCode = `UPDATE email_codes SET attempts = attempts + 1 WHERE code_id = $1`
)

// swagger:model
type EmailCode struct {
	CodeID   int    `db:"code_id"`
	Email    string `db:"email"`
	Code     string `db:"code"`
	CTime    string `db:"ctime"`
	Type     string `db:"code_type"`
	Attempts int    `db:"attempts"`
	CTimeObj time.Time
}

func (email EmailCode) Create(tx TXWrapper) error {
	_, err := tx.NamedExec(createEmailCode, email)
	return err
}

func GetLastEmailCode(db DBWrapper, email string, codeType string) (EmailCode, error) {
	var emailCode EmailCode
	err := db.Get(&emailCode, getCodeByEmail, email, codeType)
	if err != nil {
		return emailCode, err
	}

	emailCode.CTimeObj, err = time.Parse(time.RFC3339, emailCode.CTime)

	return emailCode, err
}

func (code EmailCode) IncreaseAttempts(tx TXWrapper) error {
	_, err := tx.Exec(increaseAttemptsInEmailCode, code.CodeID)
	return err
}
