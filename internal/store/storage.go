package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("record not found")
	ErrConflict          = errors.New("resource already exists")
	ErrDuplicateUsername = errors.New("username already exists")
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrWDFUQ             = errors.New("what the fck is goin on")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts    IPostStore
	Users    IUserStore
	Comments ICommentStore
	Roles    IRoleStore
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Posts: &PostStore{
			db: db,
		},
		Users: &UserStore{
			db: db,
		},
		Comments: &CommentsStore{
			db: db,
		},
		Roles: &RoleStore{
			db: db,
		},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		err1 := tx.Rollback()
		if err1 != nil {
			return err1
		}
		return err
	}

	return tx.Commit()

}
