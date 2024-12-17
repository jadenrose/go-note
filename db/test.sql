
SELECT
    n.id,
    n.title,
    b.id,
    b.sort_order,
    b.content
FROM notes n
LEFT JOIN blocks b
ON b.note_id = n.id
WHERE n.id = 1
ORDER BY sort_order ASC;