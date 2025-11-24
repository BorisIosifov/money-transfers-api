package model

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func DBConnect(config Config) (db *sqlx.DB, err error) {
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Postgres.Host, config.Postgres.Port,
		config.Postgres.User, config.Postgres.Password, config.Postgres.DBName)

	dbSqlx, err := sqlx.Connect(
		"pgx",
		connString,
	)
	if err != nil {
		return nil, err
	}

	log.Print("Postgres is ready")

	return dbSqlx, nil
}

type DBWrapper struct {
	DB *sqlx.DB
}

func (db DBWrapper) MustBegin() TXWrapper {
	tx := db.DB.MustBegin()
	return TXWrapper{TX: tx}
}

func (db DBWrapper) Get(dest interface{}, query string, args ...interface{}) error {
	log.Printf("%s; values: %v", query, args)
	return db.DB.Get(dest, query, args...)
}

type TXWrapper struct {
	TX *sqlx.Tx
}

func (tx TXWrapper) Rollback() error {
	return tx.TX.Rollback()
}

func (tx TXWrapper) Commit() error {
	return tx.TX.Commit()
}

func (tx TXWrapper) NamedExec(query string, arg interface{}) (sql.Result, error) {
	log.Printf("%s; values: %+v", query, arg)
	return tx.TX.NamedExec(query, arg)
}

func (tx TXWrapper) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	log.Printf("%s; values: %+v", query, arg)
	return tx.TX.NamedQuery(query, arg)
}

func (tx TXWrapper) Exec(query string, args ...any) (sql.Result, error) {
	log.Printf("%s; values: %+v", query, args)
	return tx.TX.Exec(query, args...)
}
