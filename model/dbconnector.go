package model

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func DBConnect(config Config) (db *sqlx.DB, err error) {
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Postgres.Host, config.Postgres.Port,
		config.Postgres.User, config.Postgres.Password, config.Postgres.DBName)

	dbSqlx, err := sqlx.Connect(
		"postgres",
		connString,
	)
	if err != nil {
		return nil, err
	}

	log.Print("Postgres is ready")

	return dbSqlx, nil
}
