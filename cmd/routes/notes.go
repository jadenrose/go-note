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

func Index(c echo.Context) error {
	notes, err := getAllPreviews()
	if err != nil {
		log.Panic(err)
		return c.NoContent(500)
	}
	if len(notes) > 0 {
		note, err := getContentByNoteId(notes[0].ID)
		if err != nil {
			log.Panic(err)
			return c.NoContent(500)
		}
		notes[0] = note
	}

	return c.Render(200, "index", notes)
}

func GetPreviewLinks(c echo.Context) error {
	notes, err := getAllPreviews()
	if err != nil {
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
		log.Panic(err)
		return c.NoContent(404)
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

	return c.Render(200, "replace-title", Note{ID: note_id, Title: title})
}

func GetNewNote(c echo.Context) error {
	return c.Render(200, "new-note", nil)
}

func PostNote(c echo.Context) error {
	title := c.FormValue("title")
	if len(title) == 0 {
		title = "Untitled Note"
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
	res, err := tx.Exec("INSERT INTO notes (title) VALUES (?);", title)
	if err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}
	if err = tx.Commit(); err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}

	return c.Render(
		200,
		"new-note-block-editor",
		Note{
			ID:     int(id),
			Title:  title,
			Blocks: []Block{{NoteID: int(id)}}},
	)
}

func ShowMoreOptions(c echo.Context) error {
	note_id, err := strconv.Atoi(c.QueryParam("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param ?note_id")
	}

	return c.Render(200, "more-options", Note{ID: note_id})
}

func HideMoreOptions(c echo.Context) error {
	note_id, err := strconv.Atoi(c.QueryParam("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param ?note_id")
	}

	return c.Render(200, "show-more-options", Note{ID: note_id})
}

func DeleteNote(c echo.Context) error {
	note_id, err := strconv.Atoi(c.Param("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :note_id")
	}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		log.Panic(err)
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Panic(err)
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(
		"DELETE FROM blocks WHERE note_id = ?;",
		note_id,
	); err != nil {
		log.Panic(err)
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec(
		"DELETE FROM notes WHERE id = ?;",
		note_id,
	); err != nil {
		log.Panic(err)
		tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		log.Panic(err)
		tx.Rollback()
		return c.NoContent(500)
	}

	return c.NoContent(200)
}

func getAllPreviews() ([]Note, error) {
	notes := []Note{}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		log.Panic(err)
		return notes, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		log.Panic(err)
		tx.Rollback()
		return notes, err
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
		return notes, err
	}
	for rows.Next() {
		n := Note{}
		err = rows.Scan(&n.ID, &n.Title)
		if err != nil {
			log.Panic(err)
			tx.Rollback()
			return notes, err
		}
		notes = append(notes, n)
	}
	if err = tx.Commit(); err != nil {
		log.Panic(err)
		tx.Rollback()
		return notes, err
	}

	return notes, nil
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
                b.note_id,
                b.sort_order,
                b.content
            FROM notes n
            LEFT JOIN blocks b
            ON b.note_id = n.id
            WHERE n.id = ?
            ORDER BY sort_order ASC;
        `,
		note_id,
	)
	if err != nil {
		tx.Rollback()
		return note, err
	}
	for rows.Next() {
		block := MaybeBlock{}
		if err = rows.Scan(
			&note.ID,
			&note.Title,
			&block.ID,
			&block.NoteID,
			&block.SortOrder,
			&block.Content,
		); err != nil {
			tx.Rollback()
			return note, err
		}

		if block.Valid() {
			note.Blocks = append(note.Blocks, block.Value())
		}
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return note, err
	}

	return note, nil
}
