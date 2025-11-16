package collab

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
func (s *Store) AddCollaborator(noteID, userID int, canEdit bool) error {
	_, err := s.db.Exec(`INSERT INTO note_collaborators (note_id, user_id, can_edit)
	VALUES ($1, $2, $3) ON CONFLICT (note_id, user_id) 
	DO UPDATE SET can_edit = EXCLUDED.can_edit`, noteID, userID, canEdit)
	return err
}

func (s *Store) RemoveCollaborator(noteID, userID int) error {
	_, err := s.db.Exec(`DELETE FROM note_collaborators WHERE note_id = $1 AND user_id = $2`,
		noteID,
		userID,
	)
	return err
}

func (s *Store) ListCollaborators(noteID int) ([]types.NoteCollaborator, error) {
	rows, err := s.db.Query(`SELECT id, note_id, user_id, can_edit, created_at
	FROM note_collaborators WHERE note_id = $1
	ORDER BY created_at DESC`, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []types.NoteCollaborator

	for rows.Next() {
		var c types.NoteCollaborator
		if err := rows.Scan(
			&c.ID,
			&c.NoteID,
			&c.UserID,
			&c.CanEdit,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Store) IsCollaborator(noteID, userID int) (bool, error) {
	var exists bool
	err := s.db.QueryRow(
		`SELECT EXISTS (
             SELECT 1
             FROM note_collaborators
             WHERE note_id = $1
               AND user_id = $2
         )`,
		noteID,
		userID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
