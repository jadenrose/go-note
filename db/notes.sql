DROP TABLE IF EXISTS notes;

CREATE TABLE
    IF NOT EXISTS notes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        modified_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        title TEXT NOT NULL
    );

INSERT INTO
    notes (title)
VALUES
    ('Lorem ipsum dolor sit amet');

DROP TABLE IF EXISTS blocks;

CREATE TABLE
    IF NOT EXISTS blocks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sort_order INTEGER NOT NULL,
        content TEXT NOT NULL,
        note_id INTEGER NOT NULL,
        FOREIGN KEY (note_id) REFERENCES notes (id)
    );

INSERT INTO
    blocks (note_id, sort_order, content)
VALUES
    (
        1,
        0,
        '1: Lorem ipsum dolor sit amet, consectetur adipiscing elit.'
    ),
    (
        1,
        1,
        '2: Lorem ipsum dolor sit amet, consectetur adipiscing elit.'
    ),
    (
        1,
        2,
        '3: Lorem ipsum dolor sit amet, consectetur adipiscing elit.'
    );