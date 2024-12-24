package routes

import (
	"database/sql"
	"errors"
	"log"
	"regexp"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Block struct {
	ID        int    `json:"id"`
	NoteID    int    `json:"note_id"`
	SortOrder int    `json:"sort_order"`
	Content   string `json:"content"`
}

type MaybeBlock struct {
	ID        sql.NullInt64
	NoteID    sql.NullInt64
	SortOrder sql.NullInt64
	Content   sql.NullString
}

func (mb MaybeBlock) Valid() bool {
	return (mb.ID.Valid &&
		mb.NoteID.Valid &&
		mb.SortOrder.Valid &&
		mb.Content.Valid)
}

func (mb MaybeBlock) Value() Block {
	return Block{
		ID:        int(mb.ID.Int64),
		NoteID:    int(mb.NoteID.Int64),
		SortOrder: int(mb.SortOrder.Int64),
		Content:   string(mb.Content.String),
	}
}

func GetNewBlock(c echo.Context) error {
	note_id, err := strconv.Atoi(c.QueryParam("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :note_id")
	}
	return c.Render(200, "block-editor--new", Block{NoteID: note_id})
}

func GetBlockEditor(c echo.Context) error {
	block_id, err := strconv.Atoi(c.Param("block_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :block_id")
	}
	block, err := getContentByBlockId(block_id)
	if err != nil {
		log.Panic(err)
		return c.NoContent(404)
	}
	return c.Render(200, "block-editor--existing", block)
}

func PostBlock(c echo.Context) error {
	var err error

	note_id, err := strconv.Atoi(c.QueryParam("note_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :note_id")
	}
	content := c.FormValue("content")
	if len(content) == 0 {
		return c.String(422, "Content cannot be empty")
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
	sort_order, err := getLastSortOrderUsed(note_id)
	if err != nil {
		return handleError()
	}
	block := Block{
		NoteID:    note_id,
		Content:   content,
		SortOrder: sort_order + 1,
	}
	res, err := agent.Exec(
		`
            INSERT INTO blocks
                (
                    note_id,
                    content,
                    sort_order
                )
            VALUES
                (
                    $1,
                    $2,
                    $3
                );
            
            UPDATE notes
            SET modified_at = CURRENT_TIMESTAMP
            WHERE id = $1;
        `,
		block.NoteID,
		block.Content,
		block.SortOrder,
	)
	if err != nil {
		return handleError()
	}
	id, err := res.LastInsertId()
	if err != nil {
		return handleError()
	}
	block.ID = int(id)
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "block-editor--afterpost", block)
}

func PutBlock(c echo.Context) error {
	block_id, err := strconv.Atoi(c.Param("block_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :block_id")
	}
	block, err := getContentByBlockId(block_id)
	if err != nil {
		log.Panic(err)
		return c.NoContent(404)
	}
	content := c.FormValue("content")
	if len(content) == 0 {
		return c.Render(422, "block", block)
	}
	if content == block.Content {
		return c.Render(200, "block", block)
	}
	block.Content = content
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
            UPDATE blocks
            SET content = $1
            WHERE id = $2;

            UPDATE notes
            SET modified_at = CURRENT_TIMESTAMP
            WHERE id = $3;
        `,
		block.Content,
		block.ID,
		block.NoteID,
	); err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return c.Render(200, "block", block)
}

func GetBlockMover(c echo.Context) error {
	block_id, err := strconv.Atoi(c.Param("block_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :block_id")
	}
	block, err := getContentByBlockId(block_id)
	if err != nil {
		log.Panic(err)
		return c.NoContent(404)
	}
	return c.Render(200, "block-mover", block)
}

func CancelBlockMover(c echo.Context) error {
	return c.NoContent(200)
}

func MoveBlock(c echo.Context) error {
	block_id, err := strconv.Atoi(c.Param("block_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :block_id")
	}
	direction := c.QueryParam("direction")
	valid, err := regexp.MatchString("up|down|top|bottom", direction)
	if !valid || err != nil {
		return c.String(400, "Missing or invalid param ?direction")
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
	rows, err := agent.Query(
		`
            SELECT id, note_id FROM blocks
            WHERE note_id = (
                SELECT note_id FROM blocks
                WHERE id = ?
                LIMIT 1
            )
            ORDER BY sort_order ASC;
        `,
		block_id,
	)
	if err != nil {
		return handleError()
	}
	i := 0
	ids := []int{}
	var note_id int
	switch direction {
	default:
		err = errors.New("didn't match any valid direction")
		return handleError()
	case "top":
		ids = append(ids, block_id)
		for rows.Next() {
			var id int
			if err = rows.Scan(&id, &note_id); err != nil {
				return handleError()
			}
			if i == 0 && id == block_id {
				agent.Rollback()
				// @TODO Use proper status code 422
				return c.String(400, "Cannot move in that direction")
			}
			if id != block_id {
				ids = append(ids, id)
			}
			i++
		}
	case "bottom":
		var index_of_block int
		for rows.Next() {
			var id int
			if err = rows.Scan(&id, &note_id); err != nil {
				return handleError()
			}
			if id == block_id {
				index_of_block = i
			} else {
				ids = append(ids, id)
			}
			i++

		}
		if index_of_block == i-1 {
			agent.Rollback()
			// @TODO Use proper status code 422
			return c.String(400, "Cannot move in that direction")
		}
		ids = append(ids, block_id)
	case "up":
		for rows.Next() {
			var id int
			if err = rows.Scan(&id, &note_id); err != nil {
				return handleError()
			}
			if i == 0 && id == block_id {
				agent.Rollback()
				// @TODO Use proper status code 422
				return c.String(400, "Cannot move in that direction")
			} else if i > 0 && id == block_id {
				ids = append(ids, ids[i-1])
				ids[i-1] = block_id
			} else if id == block_id {
				ids[i-1] = block_id
			} else {
				ids = append(ids, id)
			}
			i++
		}
	case "down":
		for rows.Next() {
			var id int
			if err = rows.Scan(&id, &note_id); err != nil {
				return handleError()
			}

			if id == block_id {
				// push twice to expand array one extra
				// 2nd one will be replaced next loop
				ids = append(ids, block_id, block_id)
			} else if len(ids) > i {
				ids[i-1] = id
			} else {
				ids = append(ids, id)
			}
			i++
		}
		if len(ids) > i {
			agent.Rollback()
			// @TODO Use proper status code 422
			return c.String(400, "Cannot move in that direction")
		}
	}
	for i, id := range ids {
		if _, err = agent.Exec(
			`
                UPDATE blocks
                SET sort_order = ?
                WHERE id = ?
            `,
			i,
			id,
		); err != nil {
			return handleError()
		}
	}
	if _, err = agent.Exec(
		`
        UPDATE notes
        SET modified_at = CURRENT_TIMESTAMP
        WHERE id = $1   
        `,
		note_id,
	); err != nil {
		return handleError()
	}

	if err = agent.Commit(); err != nil {
		return handleError()
	}

	note, err := getContentByNoteId(note_id)
	if err != nil {
		return handleError()
	}

	return c.Render(200, "blocks", note)
}

func DeleteBlock(c echo.Context) error {
	block_id, err := strconv.Atoi(c.Param("block_id"))
	if err != nil {
		return c.String(400, "Missing or invalid param :block_id")
	}
	block, err := getContentByBlockId(block_id)
	if err != nil {
		return c.NoContent(404)
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
	rows, err := agent.Query(
		`
            SELECT id, sort_order FROM blocks
            WHERE note_id = ?
            AND sort_order > ?
        `,
		block.NoteID,
		block.SortOrder,
	)
	if err != nil {
		return handleError()
	}
	for rows.Next() {
		block_to_move := Block{}
		if err = rows.Scan(
			&block_to_move.ID,
			&block_to_move.SortOrder,
		); err != nil {
			return handleError()
		}
		if _, err = agent.Exec(
			`
                UPDATE blocks
                SET sort_order = ?
                WHERE id = ?;   
            `,
			block_to_move.ID,
			block_to_move.SortOrder-1,
		); err != nil {
			return handleError()
		}
	}
	if _, err = agent.Exec(
		`
            DELETE FROM blocks
            WHERE id = $1;

            UPDATE notes
            SET modified_at = CURRENT_TIMESTAMP
            WHERE id = $2;
        `,
		block.ID,
		block.NoteID,
	); err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()

	}

	return c.NoContent(200)
}

func getContentByBlockId(block_id int) (Block, error) {
	var err error

	block := Block{}

	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() (Block, error) {
		log.Panic(err)
		agent.Rollback()
		return block, err
	}
	if err = agent.Open(); err != nil {
		return handleError()
	}
	row := agent.QueryRow(
		`
            SELECT
                id,
                note_id,
                sort_order,
                content
            FROM blocks
                WHERE id = ?;
        `,
		block_id,
	)
	if err = row.Scan(
		&block.ID,
		&block.NoteID,
		&block.SortOrder,
		&block.Content,
	); err != nil {
		return handleError()
	}
	if err = agent.Commit(); err != nil {
		return handleError()
	}

	return block, nil
}

func getLastSortOrderUsed(note_id int) (int, error) {
	var err error

	if agent == nil {
		agent = NewDBAgent()
	}
	handleError := func() (int, error) {
		log.Panic(err)
		return -1, err
	}

	if err = agent.Open(); err != nil {
		return handleError()
	}
	defer agent.Close()
	var sort_order sql.NullInt64
	row := agent.QueryRow(
		`
            SELECT MAX(sort_order)
            FROM blocks
            WHERE note_id = ?;
        `,
		note_id,
	)
	if err = row.Scan(&sort_order); err != nil {
		return handleError()
	}
	if !sort_order.Valid {
		sort_order.Int64 = 0
	}

	return int(sort_order.Int64), nil
}
