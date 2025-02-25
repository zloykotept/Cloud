package store

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	Users interface {
		Create(user *User) error
		Update(user *User) error
		UpdateField(id int, field string, val any) error
		Delete(id int) error
		GetByName(name string) (*User, error)
		GetById(id int) (*User, error)
		GetPage(page int, size int) (*[]User, error)
	}
	Files interface {
		Create(file *File) (*File, error)
		UpdateField(id uuid.UUID, owner int, field string, val any) error
		Delete(id uuid.UUID, owner int) error
		GetById(id uuid.UUID, owner int) (*File, error)
		GetPage(owner int, page int, size int) (*[]File, error)
		GetPagePublic(page int, size int) (*[]File, error)
		GetOccupiedSpace(owner int) (float32, error)
	}
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		Users: &UsersStorage{db},
		Files: &FilesStorage{db},
	}
}
