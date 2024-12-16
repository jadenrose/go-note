package notes

import (
	"database/sql"
)

type Block struct {
	ID        int    `json:"id"`
	NoteID    int    `json:"note_id"`
	SortOrder int    `json:"sort_order"`
	Content   string `json:"content"`
}

type Note struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Blocks []Block `json:"content"`
}

func GetNote(id int) (Note, error) {
	note := Note{}
	block := Block{}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return note, err
	}
	defer db.Close()
	rows, err := db.Query(
		`
            SELECT
                n.id,
                n.title,
                b.id,
                b.sort_order,
                b.content
            FROM notes n, blocks b
                WHERE b.note_id = n.id
                AND n.id = ?
            ORDER BY b.sort_order ASC;
        `,
		id,
	)
	if err != nil {
		return note, err
	}
	for rows.Next() {
		err = rows.Scan(
			&note.ID,
			&note.Title,
			&block.ID,
			&block.SortOrder,
			&block.Content,
		)
		if err != nil {
			return note, err
		}
		note.Blocks = append(note.Blocks, block)
	}

	return note, nil
}

func GetBlock(id int) (Block, error) {
	block := Block{}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return block, err
	}
	defer db.Close()
	row := db.QueryRow(
		`
            SELECT
                id,
                note_id,
                sort_order,
                content
            FROM blocks
                WHERE id = ?;
        `,
		id,
	)
	err = row.Scan(
		&block.ID,
		&block.NoteID,
		&block.SortOrder,
		&block.Content,
	)
	if err != nil {
		return block, err
	}

	return block, nil
}

func GetAllNotes() ([]Note, error) {
	notes := []Note{}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return notes, err
	}
	defer db.Close()
	rows, err := db.Query(
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

func UpdateNote(note Note) error {
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(
		`
            UPDATE notes
            SET
                title = ?
            WHERE id = ?;
        `,
		note.Title,
		note.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func AddBlock(note_id int, content string) (Block, error) {
	block := Block{NoteID: note_id, Content: content}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return block, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return block, err
	}
	sort_order, err := GetLastSortOrder(block)
	if err != nil {
		return block, err
	}
	sort_order++
	block.SortOrder = sort_order
	row, err := tx.Exec(
		`
            INSERT INTO blocks (note_id, content, sort_order)
            VALUES (?, ?, ?);
        `,
		note_id,
		content,
		sort_order,
	)
	if err != nil {
		return block, err
	}
	id, err := row.LastInsertId()
	if err != nil {
		return block, err
	}
	block.ID = int(id)
	err = tx.Commit()
	return block, err
}

func UpdateBlock(block Block) error {
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(
		`
            UPDATE blocks
            SET
                content = ?
            WHERE id = ?;
        `,
		block.Content,
		block.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

func MoveBlock(block Block, new_sort_order int) error {
	if new_sort_order == block.SortOrder {
		return nil
	}
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if new_sort_order < block.SortOrder {
		rows, err := tx.Query(
			`
                SELECT id, sort_order FROM blocks
                WHERE note_id = ?
                AND sort_order >= ?
                AND sort_order < ?;
            `,
			block.NoteID,
			new_sort_order,
			block.SortOrder,
		)
		if err != nil {
			return err
		}
		var id, sort_order int
		for rows.Next() {
			err = rows.Scan(&id, &sort_order)
			if err != nil {
				return err
			}
			_, err = tx.Exec("UPDATE blocks SET sort_order = ? WHERE id = ?;", sort_order+1, id)
			if err != nil {
				return err
			}
		}
	} else {
		rows, err := tx.Query(
			`
                SELECT id, sort_order FROM blocks
                WHERE note_id = ?
                AND sort_order <= ?
                AND sort_order > ?;
            `,
			block.NoteID,
			new_sort_order,
			block.SortOrder,
		)
		if err != nil {
			return err
		}
		var id, sort_order int
		for rows.Next() {
			err = rows.Scan(&id, &sort_order)
			if err != nil {
				return err
			}
			_, err = tx.Exec("UPDATE blocks SET sort_order = ? WHERE id = ?;", sort_order-1, id)
			if err != nil {
				return err
			}
		}
	}
	_, err = tx.Exec("UPDATE blocks SET sort_order = ? WHERE id = ?;", new_sort_order, block.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func DeleteBlock(block Block) error {
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	rows, err := tx.Query(
		`
            SELECT id, sort_order FROM blocks
            WHERE note_id = ?
            AND sort_order > ?
        `,
		block.NoteID,
		block.SortOrder,
	)
	if err != nil {
		return err
	}
	for rows.Next() {
		block_to_move := Block{}
		if err = rows.Scan(&block_to_move.ID, &block_to_move.SortOrder); err != nil {
			return err
		}
		_, err = tx.Exec(
			`
                UPDATE blocks
                SET sort_order = ?
                WHERE id = ?;   
            `,
			block_to_move.ID,
			block_to_move.SortOrder-1,
		)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(`DELETE FROM blocks WHERE id = ?`, block.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func GetLastSortOrder(block Block) (int, error) {
	db, err := sql.Open("sqlite", "./db/notes.db")
	if err != nil {
		return -1, err
	}
	defer db.Close()
	var last_sort_order int
	row := db.QueryRow(
		`
            SELECT MAX(sort_order)
            FROM blocks
            WHERE note_id = ?
            AND id != ?;
        `,
		block.NoteID,
		block.ID,
	)
	err = row.Scan(&last_sort_order)
	return last_sort_order, err
}
