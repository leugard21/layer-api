package types

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

type Note struct {
	ID         int       `json:"id"`
	OwnerID    int       `json:"ownerId"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	IsArchived bool      `json:"isArchived"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type NoteCollaborator struct {
	ID        int       `json:"id"`
	NoteID    int       `json:"noteId"`
	UserID    int       `json:"userId"`
	CanEdit   bool      `json:"canEdit"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserStore interface {
	CreateUser(User) (int, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByID(id int) (*User, error)
}

type NoteStore interface {
	CreateNote(note Note) (int, error)
	GetNoteByID(id int) (*Note, error)
	ListNotesByOwner(ownerID int) ([]Note, error)
	UpdateNote(note Note) error
	ArchiveNote(id int, ownerID int) error
}

type CollaboratorStore interface {
	AddCollaborator(noteID, userID int, canEdit bool) error
	RemoveCollaborator(noteID, userID int) error
	ListCollaborators(noteID int) ([]NoteCollaborator, error)
	IsCollaborator(noteID, userID int) (bool, error)
}

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,min=3,max=30"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=130"`
}

type LoginPayload struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type RefreshPayload struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type CreateNotePayload struct {
	Title   string `json:"title" validate:"max=200"`
	Content string `json:"content" validate:"max=100000"`
}

type UpdateNotePayload struct {
	Title   *string `json:"title,omitempty" validate:"omitempty,max=200"`
	Content *string `json:"content,omitempty" validate:"omitempty,max=100000"`
}

type AddCollaboratorPayload struct {
	UserID  int   `json:"userId" validate:"required"`
	CanEdit *bool `json:"canEdit,omitempty"`
}
