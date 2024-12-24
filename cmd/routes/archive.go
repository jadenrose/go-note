package routes

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ArchivePreview struct {
	ID         int
	Title      string
	BlockCount int
}

type MaybeArchivePreview struct {
	ID         *sql.NullInt64
	Title      *sql.NullString
	BlockCount *sql.NullInt64
}

func (mp MaybeArchivePreview) Valid() bool {
	return (mp.ID.Valid &&
		mp.Title.Valid &&
		mp.BlockCount.Valid)
}

func (mp MaybeArchivePreview) Value() ArchivePreview {
	return ArchivePreview{
		ID:         int(mp.ID.Int64),
		Title:      string(mp.Title.String),
		BlockCount: int(mp.BlockCount.Int64),
	}
}

func GetArchiveList(c echo.Context) error {
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
	rows, err := agent.Query(
		`
        SELECT DISTINCT n.id, n.title, COUNT(b.id)
        FROM notes_archive n
        LEFT JOIN blocks_archive b
        ON b.note_id = n.id
        GROUP BY n.id;
        `,
	)
	var notes []ArchivePreview
	for rows.Next() {
		note := MaybeArchivePreview{}

		if err = rows.Scan(
			&note.ID,
			&note.Title,
			&note.BlockCount,
		); err != nil {
			return handleError()
		}

		if note.Valid() {
			notes = append(notes, note.Value())
		}
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "archive", notes)
}

func GetArchivedNote(c echo.Context) error {
	var err error
	archived_note_id, err := strconv.Atoi(c.Param("archived_note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :archived_note_id")
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
	rows, err := agent.Query(
		`
        SELECT
            n.id,
            n.title,
            b.id,
            b.note_id,
            b.content,
            b.sort_order
        FROM notes_archive n, blocks_archive b
        WHERE b.note_id = n.id
        AND b.note_id = ?
        ORDER BY b.sort_order;
        `,
		archived_note_id,
	)
	if err != nil {
		return handleError()
	}
	note := Note{}
	for rows.Next() {
		block := MaybeBlock{}
		if err = rows.Scan(
			&note.ID,
			&note.Title,
			&block.ID,
			&block.NoteID,
			&block.Content,
			&block.SortOrder,
		); err != nil {
			return handleError()
		}

		if block.Valid() {
			note.Blocks = append(note.Blocks, block.Value())
		}
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "readonly-main", note)
}

func RestoreArchivedNote(c echo.Context) error {
	var err error
	archived_note_id, err := strconv.Atoi(c.Param("archived_note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :archived_note_id")
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

	res, err := agent.Exec(
		`
        INSERT INTO notes (title, modified_at)
        SELECT title, CURRENT_TIMESTAMP FROM notes_archive
        WHERE id = ?;
        `,
		archived_note_id,
	)
	if err != nil {
		return handleError()
	}
	note_id, err := res.LastInsertId()
	if err != nil {
		return handleError()
	}

	if _, err = agent.Exec(
		`
        INSERT INTO blocks (content, sort_order, note_id)
        SELECT content, sort_order, $1 FROM blocks_archive
        WHERE note_id = $2;
        `,
		int(note_id),
		archived_note_id,
	); err != nil {
		return handleError()
	}

	if _, err = agent.Exec(
		`
        DELETE FROM notes_archive
        WHERE id = $1;

        DELETE FROM blocks_archive
        WHERE note_id = $1;
        `,
		archived_note_id,
	); err != nil {
		return handleError()
	}
	if err = archiveOldNotes(); err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	note, err := getContentByNoteId(int(note_id))

	return c.Render(200, "restored-note", note)

}

func ClearArchive(c echo.Context) error {
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
	if _, err = agent.Exec(
		`
        DELETE FROM blocks_archive;
        DELETE FROM notes_archive;
        `,
	); err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "empty-archive", nil)
}

func archiveOldNotes() error {
	var err error

	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() error {
		log.Panic(err)
		agent.Rollback()
		return err
	}
	if err = agent.Open(); err != nil {
		return handleError()
	}
	defer agent.Close()
	rows, err := agent.Query(
		`
        SELECT id FROM (
            SELECT
                *,
                row_number() OVER (
                    ORDER BY modified_at DESC
                ) as rank
            FROM notes
        )
        WHERE rank > 20;
        `,
	)
	if err != nil {
		return handleError()
	}
	ids_to_archive := []int{}
	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return handleError()
		}
		ids_to_archive = append(ids_to_archive, id)
	}

	for _, id := range ids_to_archive {
		if _, err = archiveNoteById(id); err != nil {
			return handleError()
		}
	}

	return nil
}

func archiveNoteById(note_id int) (int, error) {
	var err error

	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() (int, error) {
		log.Panic(err)
		agent.Rollback()
		return -1, err
	}
	if err = agent.Open(); err != nil {
		return handleError()
	}
	res, err := agent.Exec(
		`
        INSERT INTO notes_archive (
            title,
            created_at,
            modified_at,
            archived_at
        )
        SELECT
            title,
            created_at,
            modified_at,
            CURRENT_TIMESTAMP
        FROM notes
        WHERE id = $1;
        `,
		note_id,
	)
	if err != nil {
		return handleError()
	}

	archived_note_id, err := res.LastInsertId()
	if err != nil {
		return handleError()
	}
	if _, err = agent.Exec(
		`
        INSERT INTO blocks_archive (
            note_id,
            content,
            sort_order
        )
        SELECT
            $1,
            content,
            sort_order
        FROM blocks
        WHERE note_id = $2;
        `,
		int(archived_note_id),
		note_id,
	); err != nil {
		return handleError()
	}

	if _, err = agent.Exec(
		`
        DELETE FROM notes
        WHERE id = $1;
        
        DELETE FROM blocks
        WHERE note_id = $1;
        `,
		note_id,
	); err != nil {
		return handleError()
	}

	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return int(archived_note_id), nil
}
