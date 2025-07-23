package model

const (
	createSessionQuery = "INSERT INTO sessions (session_id, user_id, data) VALUES (:session_id, :user_id, :data)"

	getSessionQuery = "SELECT session_id, user_id, data, ctime from sessions WHERE session_id = $1"
)

// swagger:model
type Session struct {
	SessionID string  `db:"session_id"`
	UserID    NullInt `db:"user_id"`
	Data      string  `db:"data"`
	CTime     string  `db:"ctime"`
}

func (session Session) Create(tx TXWrapper) error {
	_, err := tx.NamedExec(createSessionQuery, session)
	return err
}

func GetSession(db DBWrapper, sessionID string) (Session, error) {
	var session Session
	err := db.Get(&session, getSessionQuery, sessionID)
	return session, err
}
