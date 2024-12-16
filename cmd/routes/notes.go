package routes

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Note struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Blocks []Block `json:"content"`
}

func GetPreviewLinks(c echo.Context) error {
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		log.Panic(err)
		return c.NoContent(500)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}
	rows, err := tx.Query(
		`
            SELECT id, title FROM notes
            ORDER BY modified_at DESC;
        `,
	)
	if err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}
	notes := []Note{}
	for rows.Next() {
		n := Note{}
		err = rows.Scan(&n.ID, &n.Title)
		if err != nil {
			log.Panic(err)
			tx.Rollback()
			return c.NoContent(500)
		}
		notes = append(notes, n)
	}
	if err = tx.Commit(); err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}

	return c.Render(200, "preview-links", notes)
}

func GetNoteContent(c echo.Context) error {
	note_id, err := strconv.Atoi(c.Param("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :note_id")
	}
	note, err := getContentByNoteId(note_id)
	if err != nil {
		return c.NoContent(400)
	}

	return c.Render(200, "note-content", note)
}

func GetTitleEditor(c echo.Context) error {
	note_id, err := strconv.Atoi(c.Param("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :note_id")
	}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		log.Panic(err)
		return c.NoContent(500)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}
	row := tx.QueryRow("SELECT id, title FROM notes WHERE id = ?", note_id)
	note := Note{}
	if err = row.Scan(&note.ID, &note.Title); err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}
	if err = tx.Commit(); err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}

	return c.Render(200, "title-editor", note)
}

func PutTitle(c echo.Context) error {
	note_id, err := strconv.Atoi(c.Param("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :note_id")
	}
	title := c.FormValue("title")
	if len(title) == 0 {
		return c.String(422, "Title cannot be empty")
	}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		log.Panic(err)
		return c.NoContent(500)
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}
	if _, err := tx.Exec(
		`
            UPDATE notes
            SET title = ?
            WHERE id = ?
        `,
		title,
		note_id,
	); err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}
	if err = tx.Commit(); err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}

	return c.Render(200, "title", Note{ID: note_id, Title: title})
}

func getContentByNoteId(note_id int) (Note, error) {
	note := Note{}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return note, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return note, err
	}
	rows, err := tx.Query(
		`
            SELECT
                n.id,
                n.title,
                b.id,
                b.sort_order,
                b.content
            FROM notes n, blocks b
            WHERE n.id = ?
            AND b.note_id = n.id
            ORDER BY b.sort_order ASC
        `,
		note_id,
	)
	if err != nil {
		tx.Rollback()
		return note, err
	}
	for rows.Next() {
		block := Block{}
		if err = rows.Scan(
			&note.ID,
			&note.Title,
			&block.ID,
			&block.SortOrder,
			&block.Content,
		); err != nil {
			tx.Rollback()
			return note, err
		}
		note.Blocks = append(note.Blocks, block)
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return note, err
	}

	return note, nil
}
