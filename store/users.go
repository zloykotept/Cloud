package store

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type User struct {
	Id          int     `db:"id"`
	Username    string  `db:"name"`
	Password    []byte  `db:"password"`
	Permissions uint8   `db:"permissions"`
	Space       float32 `db:"space"`
}

func NewUser(n string, pass []byte, per uint8, s float32) *User {
	return &User{
		Username:    n,
		Password:    pass,
		Permissions: per,
		Space:       s,
	}
}

type UsersStorage struct {
	db *sqlx.DB
}

func (us *UsersStorage) Create(u *User) error {
	_, err := us.db.NamedExec(`INSERT INTO users (name, password, permissions, space) VALUES (:name, :password, :permissions, :space)`, u)
	return err
}

func (us *UsersStorage) Update(u *User) error {
	_, err := us.db.NamedExec(`UPDATE users SET name = :name, password = :password, permissions = :permissions, space = :space WHERE id = :id`, u)
	return err
}

func (us *UsersStorage) UpdateField(id int, field string, val any) error {
	q := fmt.Sprintf("UPDATE users SET %s = $1 WHERE id = $2", field)
	_, err := us.db.Exec(q, val, id)
	return err
}

func (us *UsersStorage) Delete(id int) error {
	_, err := us.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}

func (us *UsersStorage) GetByName(n string) (*User, error) {
	u := &User{}
	err := us.db.Get(u, `SELECT * FROM users WHERE name = $1`, n)
	return u, err
}

func (us *UsersStorage) GetById(id int) (*User, error) {
	u := &User{}
	err := us.db.Get(u, `SELECT * FROM users WHERE id = $1`, id)
	return u, err
}

func (us *UsersStorage) GetPage(page int, size int) (*[]User, error) {
	u := &[]User{}
	err := us.db.Select(u, `SELECT * FROM users WHERE permissions = 0 LIMIT $1 OFFSET $2`, size, (page-1)*size)
	return u, err
}
