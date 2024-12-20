package routes

import (
	"database/sql"
	"errors"
)

var agent *DBAgent

type DBAgent struct {
	Path string
	DB   *sql.DB
	Tx   *sql.Tx
}

func NewDBAgent() *DBAgent {
	return &DBAgent{}
}

func (agent *DBAgent) Open() error {
	agent.Path = "./db/notes.db"
	db, err := sql.Open("sqlite", agent.Path)
	if err != nil {
		return err
	}
	agent.DB = db
	return err
}

func (agent *DBAgent) Close() error {
	if agent.DB == nil {
		return errors.New("cannot close: db is not open")
	}

	err := agent.DB.Close()

	if err == nil {
		agent.DB = nil
	}

	return err
}

func (agent *DBAgent) Rollback() error {
	if agent.Tx == nil {
		return nil
	}

	err := agent.Tx.Rollback()

	if err == nil {
		agent.Tx = nil
	}

	return err
}

func (agent *DBAgent) Commit() error {
	if agent.Tx == nil {
		return nil
	}

	err := agent.Tx.Commit()

	if err == nil {
		agent.Tx = nil
	}

	return err
}

func (agent *DBAgent) Exec(query string, args ...any) (sql.Result, error) {
	if agent.Tx == nil {
		tx, err := agent.DB.Begin()
		if err != nil {
			return nil, err
		}
		agent.Tx = tx
	}
	res, err := agent.Tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (agent *DBAgent) Query(query string, args ...any) (*sql.Rows, error) {
	if agent.Tx == nil {
		tx, err := agent.DB.Begin()
		if err != nil {
			return nil, err
		}
		agent.Tx = tx
	}
	rows, err := agent.Tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return rows, err
}

func (agent *DBAgent) QueryRow(query string, args ...any) *sql.Row {
	if agent.Tx == nil {
		tx, err := agent.DB.Begin()
		if err != nil {
			return nil
		}
		agent.Tx = tx
	}
	row := agent.Tx.QueryRow(query, args...)

	return row
}
