package store

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type FilesStorage struct {
	db *sqlx.DB
}

type File struct {
	Id        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Size      float32   `db:"size"`
	Update    string    `db:"update"`
	Public    bool      `db:"public"`
	Favourite bool      `db:"favourite"`
	Owner     int       `db:"owner_id"`
	Dir       uuid.UUID `db:"dir_id"`
}

func NewFile(id uuid.UUID, name string, size float32, owner int) *File {
	return &File{
		Id:        id,
		Name:      name,
		Size:      size,
		Owner:     owner,
		Public:    false,
		Favourite: false,
	}
}

func (fs *FilesStorage) Create(f *File) (*File, error) {
	file := &File{}
	err := fs.db.Get(file, `INSERT INTO files (id, name, size, update, public, favourite, owner_id) VALUES ($4, $1, $2, NOW(), false, false, $3) RETURNING id, name, size, update, public, favourite, owner_id, dir_id`, f.Name, f.Size, f.Owner, f.Id)
	return file, err
}

func (fs *FilesStorage) Delete(id uuid.UUID, ow int) error {
	res, err := fs.db.Exec(`DELETE FROM files WHERE id = $1 AND owner_id = $2`, id, ow)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("ERR_FORBIDDEN")
	}
	return nil
}

func (fs *FilesStorage) GetPage(id int, page int, size int) (*[]File, error) {
	files := &[]File{}
	err := fs.db.Select(files, `SELECT * FROM files WHERE owner_id = $1 AND public = $4 LIMIT $2 OFFSET $3`, id, size, (page-1)*size, false)
	return files, err
}

func (fs *FilesStorage) GetPagePublic(page int, size int) (*[]File, error) {
	files := &[]File{}
	err := fs.db.Select(files, `SELECT * FROM files WHERE public = $1 LIMIT $2 OFFSET $3`, true, size, (page-1)*size)
	return files, err
}

func (fs *FilesStorage) UpdateField(id uuid.UUID, ow int, field string, val any) error {
	q := fmt.Sprintf(`UPDATE files SET %s = $1 WHERE id = $2 AND owner_id = $3`, field)
	res, err := fs.db.Exec(q, val, id, ow)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("ERR_FORBIDDEN")
	}
	return nil
}

func (fs *FilesStorage) GetById(id uuid.UUID, ow int) (*File, error) {
	file := &File{}
	err := fs.db.Get(file, `SELECT * FROM files WHERE (id = $1 AND owner_id = $2) OR (id = $1 AND public = $3)`, id, ow, true)
	return file, err
}

func (fs *FilesStorage) GetOccupiedSpace(owner int) (float32, error) {
	var i float32
	err := fs.db.Get(&i, `SELECT COALESCE(SUM(size), 0) FROM files WHERE owner_id = $1`, owner)
	return i, err
}
