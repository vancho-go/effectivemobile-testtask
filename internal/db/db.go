package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/vancho-go/effectivemobile-testtask/internal/logger"
)

type Database struct {
	Conn *sql.DB
}

var DB = Database{}

func Initialize(username, password, database, host, port string) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, username, password, database)
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	DB.Conn = conn
	err = DB.Conn.Ping()
	if err != nil {
		return err
	}
	logger.Log.Info("Connected to database")
	return nil
}
