package routes

import (
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"
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

	if _, err = db.Exec(
		`
        PRAGMA journal_mode = WAL;
        PRAGMA synchronous = normal;
        PRAGMA journal_size_limit = 6144000;
        `,
	); err != nil {
		return err
	}

	if _, err := db.Exec(
		`
        CREATE TABLE IF NOT EXISTS notes (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            modified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            title TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS blocks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            sort_order INTEGER NOT NULL,
            content TEXT NOT NULL,
            note_id INTEGER NOT NULL,
            FOREIGN KEY (note_id) REFERENCES notes (id)
        );

        CREATE VIRTUAL TABLE IF NOT EXISTS quick_search
        USING fts5(note_id UNINDEXED, title, content, tokenize="trigram");

        CREATE TRIGGER IF NOT EXISTS add_note_to_quick_search
        AFTER INSERT ON notes
            BEGIN
                INSERT INTO quick_search (note_id, title)
                VALUES (NEW.id, NEW.title);
            END;

        CREATE TRIGGER IF NOT EXISTS update_note_in_quick_search
        AFTER UPDATE OF title ON notes
            BEGIN
                UPDATE quick_search
                SET title = NEW.title
                WHERE note_id = NEW.id;
            END;

        CREATE TRIGGER IF NOT EXISTS remove_note_from_quick_search
        AFTER DELETE ON notes
            BEGIN
                DELETE FROM quick_search
                WHERE note_id = OLD.id;
            END;

        CREATE TRIGGER IF NOT EXISTS add_block_to_quick_search
        AFTER INSERT ON blocks
            BEGIN
                UPDATE quick_search
                SET (content) = (
                    SELECT
                        group_concat(content, ' | ')
                    FROM blocks
                    WHERE note_id = NEW.note_id
                )
                WHERE note_id = NEW.note_id;
            END;

        CREATE TRIGGER IF NOT EXISTS update_block_in_quick_search
        AFTER UPDATE OF content ON blocks
            BEGIN
                UPDATE quick_search
                SET (content) = (
                    SELECT
                        group_concat(content, ' | ')
                    FROM blocks
                    WHERE note_id = NEW.note_id
                )
                WHERE note_id = NEW.note_id;
            END;

        CREATE TABLE IF NOT EXISTS notes_archive (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            created_at DATETIME NOT NULL,
            modified_at DATETIME NOT NULL,
            archived_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            title TEXT NOT NULL
        );
        CREATE TABLE IF NOT EXISTS blocks_archive (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            sort_order INTEGER NOT NULL,
            content TEXT NOT NULL,
            note_id INTEGER NOT NULL,
            FOREIGN KEY (note_id) REFERENCES notes_archive (id)
        );
        `,
	); err != nil {
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
