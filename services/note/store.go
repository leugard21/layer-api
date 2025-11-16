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
	panic("not implemented")
}

func (s *Store) GetNoteByID(id int) (*types.Note, error) {
	panic("not implemented")
}

func (s *Store) ListNotesByOwner(ownerID int) ([]types.Note, error) {
	panic("not implemented")
}

func (s *Store) UpdateNote(note types.Note) error {
	panic("not implemented")
}

func (s *Store) ArchiveNote(id int, ownerID int) error {
	panic("not implemented")
}
