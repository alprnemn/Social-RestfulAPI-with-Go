package store

import (
	"context"
	"database/sql"
	"social/internal/types"
)

type IRoleStore interface {
	GetByName(context.Context, string) (*types.Role, error)
}

type RoleStore struct {
	db *sql.DB
}

func (store *RoleStore) GetByName(ctx context.Context, rolename string) (*types.Role, error) {

	query := `SELECT * FROM roles WHERE name=$1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	role := &types.Role{}

	err := store.db.QueryRowContext(ctx, query, rolename).Scan(
		&role.ID,
		&role.Name,
		&role.Level,
		&role.Description,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return role, nil

}
