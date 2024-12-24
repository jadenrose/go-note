package routes

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Note struct {
	ID         int
	Title      string
	CreatedAt  string
	ModifiedAt string
	Blocks     []Block
}

type MaybeNote struct {
	ID         sql.NullInt64
	Title      sql.NullString
	CreatedAt  sql.NullString
	ModifiedAt sql.NullString
	Blocks     []Block
}

func (mn MaybeNote) Valid() bool {
	return mn.ID.Valid && mn.Title.Valid
}

func (mn MaybeNote) Value() Note {
	return Note{
		ID:     int(mn.ID.Int64),
		Title:  mn.Title.String,
		Blocks: mn.Blocks,
	}
}

func Index(c echo.Context) error {
	var err error

	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() error {
		log.Panic(err)
		agent.Rollback()
		return c.NoContent(500)
	}

	if err = agent.Open(); err != nil {
		return handleError()
	}
	defer agent.Close()

	notes, err := getAllPreviews()
	if err != nil {
		return handleError()
	}
	if len(notes) > 0 {
		note, err := getContentByNoteId(notes[0].ID)
		if err != nil {
			return handleError()
		}
		notes[0] = note

		return c.Render(200, "index", notes)
	}

	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "blank-index", nil)

}

func GetPreviewLinks(c echo.Context) error {
	var err error

	agent := NewDBAgent()
	handleError := func() error {
		log.Panic(err)
		agent.Rollback()
		return c.NoContent(500)
	}
	if err = agent.Open(); err != nil {
		return handleError()
	}
	defer agent.Close()
	notes, err := getAllPreviews()
	if err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "preview-links", notes)
}

func GetNoteContent(c echo.Context) error {
	var err error

	if agent == nil {
		agent = NewDBAgent()
	}
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
	var err error

	note_id, err := strconv.Atoi(c.Param("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :note_id")
	}
	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() error {
		log.Panic(err)
		agent.Rollback()
		return c.NoContent(500)
	}
	if err := agent.Open(); err != nil {
		return handleError()
	}
	defer agent.Close()
	row := agent.QueryRow("SELECT id, title FROM notes WHERE id = ?", note_id)
	note := Note{}
	if err = row.Scan(&note.ID, &note.Title); err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "title-editor", note)
}

func PutTitle(c echo.Context) error {
	var err error

	note_id, err := strconv.Atoi(c.Param("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :note_id")
	}
	title := c.FormValue("title")
	if len(title) == 0 {
		return c.String(422, "Title cannot be empty")
	}
	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() error {
		log.Panic(err)
		agent.Rollback()
		return c.NoContent(500)
	}
	if err = agent.Open(); err != nil {
		return handleError()
	}
	defer agent.Close()
	if _, err := agent.Exec(
		`
            UPDATE notes
            SET
                title = ?
                modified_at = CURRENT_TIMESTAMP
            WHERE id = ?;
        `,
		title,
		note_id,
	); err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "replace-title", Note{ID: note_id, Title: title})
}

func GetNewNote(c echo.Context) error {
	return c.Render(200, "new-note", nil)
}

func PostNote(c echo.Context) error {
	var err error

	title := c.FormValue("title")
	if len(title) == 0 {
		title = "Untitled Note"
	}
	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() error {
		log.Panic(err)
		agent.Rollback()
		return c.NoContent(500)
	}
	if err = agent.Open(); err != nil {
		return handleError()
	}
	defer agent.Close()
	res, err := agent.Exec("INSERT INTO notes (title) VALUES (?);", title)
	if err != nil {
		return handleError()
	}
	new_note_id, err := res.LastInsertId()
	if err != nil {
		return handleError()
	}
	if err = archiveOldNotes(); err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(
		200,
		"new-note-block-editor",
		Note{
			ID:     int(new_note_id),
			Title:  title,
			Blocks: []Block{{NoteID: int(new_note_id)}}},
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
	var err error

	note_id, err := strconv.Atoi(c.Param("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :note_id")
	}
	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() error {
		log.Panic(err)
		agent.Rollback()
		return c.NoContent(500)
	}
	if err = agent.Open(); err != nil {
		return handleError()
	}
	defer agent.Close()

	if _, err = archiveNoteById(note_id); err != nil {
		return handleError()
	}

	row := agent.QueryRow(
		`
        SELECT id FROM notes
        ORDER BY modified_at DESC LIMIT 1;
        `)
	next_note_id := sql.NullInt64{}
	if err = row.Scan(&next_note_id); err != nil {
		if err = agent.Commit(); err != nil {
			return handleError()
		}
		return c.Render(200, "blank-note-oob", nil)
	}
	if !next_note_id.Valid {
		if err = agent.Commit(); err != nil {
			return handleError()
		}
		return c.Render(200, "blank-note-oob", nil)
	}
	next_note, err := getContentByNoteId(int(next_note_id.Int64))
	if err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "note-oob", next_note)

}

func getAllPreviews() ([]Note, error) {
	notes := []Note{}
	if agent == nil {
		agent = NewDBAgent()
	}
	rows, err := agent.Query(
		`
            SELECT id, title FROM notes
            ORDER BY modified_at DESC;
        `,
	)
	if err != nil {
		return notes, err
	}
	for rows.Next() {
		n := Note{}
		err = rows.Scan(&n.ID, &n.Title)
		if err != nil {
			return notes, err
		}
		notes = append(notes, n)
	}

	return notes, nil
}

func getContentByNoteId(note_id int) (Note, error) {
	note := Note{}
	if agent == nil {
		agent = NewDBAgent()
	}
	var err error
	handleError := func() (Note, error) {
		agent.Rollback()
		return note, err
	}
	if err = agent.Open(); err != nil {
		return handleError()
	}
	rows, err := agent.Query(
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
		return handleError()
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
			return handleError()
		}

		if block.Valid() {
			note.Blocks = append(note.Blocks, block.Value())
		}
	}

	return note, nil
}
