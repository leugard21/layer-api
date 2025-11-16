package note

import (
	"database/sql"
	"layer-api/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateNote(note types.Note) (int, error) {
	var id int

	err := s.db.QueryRow(`INSERT INTO notes (owner_id, title, content)
	VALUES ($1, $2, $3) RETURNING id`, note.OwnerID, note.Title, note.Content).Scan(&id)
	if err != nil {
		return 0, nil
	}

	return id, nil
}

func (s *Store) GetNoteByID(id int) (*types.Note, error) {
	row := s.db.QueryRow(`SELECT id, owner_id, title, content, is_archived, created_at, updated_at
	FROM notes WHERE id = $1 LIMIT 1`, id)

	var n types.Note
	err := row.Scan(
		&n.ID,
		&n.OwnerID,
		&n.Title,
		&n.Content,
		&n.IsArchived,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &n, nil
}

func (s *Store) ListNotesByOwner(ownerID int) ([]types.Note, error) {
	rows, err := s.db.Query(`SELECT id, owner_id, title, content, is_archived, created_at, updated_at
	FROM notes WHERE owner_id = $1 AND is_archived = FALSE ORDER BY updated_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []types.Note

	for rows.Next() {
		var n types.Note
		if err := rows.Scan(
			&n.ID,
			&n.OwnerID,
			&n.Title,
			&n.Content,
			&n.IsArchived,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (s *Store) UpdateNote(note types.Note) error {
	res, err := s.db.Exec(`UPDATE notes SET title = $1, content = $2, updated_at = NOW()
	WHERE id = $3 AND owner_id = $4`, note.Title, note.Content, note.ID, note.OwnerID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Store) ArchiveNote(id int, ownerID int) error {
	res, err := s.db.Exec(`UPDATE notes SET is_archived = TRUE, updated_at = NOW() 
	WHERE id = $1 AND owner_id = $2 AND is_archived = FALSE`, id, ownerID)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Store) UpdateNoteContent(id int, content string) error {
	res, err := s.db.Exec(`UPDATE notes SET content = $1, updated_at = NOW()
         WHERE id = $2`,
		content,
		id,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
