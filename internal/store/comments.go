package store

import (
	"context"
	"database/sql"
	"social/internal/types"
)

type ICommentStore interface {
	GetByPostId(ctx context.Context, postId int64) ([]types.Comment, error)
}

type CommentsStore struct {
	db *sql.DB
}

func (store *CommentsStore) GetByPostId(ctx context.Context, postId int64) ([]types.Comment, error) {

	query := `
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, users.username, users.id 
		FROM comments c
		JOIN users ON users.id = c.user_id
		WHERE c.post_id = $1 
		ORDER BY c.created_at DESC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := store.db.QueryContext(ctx, query, postId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []types.Comment{}

	for rows.Next() {
		var c types.Comment
		c.User = types.User{}

		err := rows.Scan(
			&c.ID,
			&c.PostId,
			&c.UserId,
			&c.Content,
			&c.CreatedAt,
			&c.User.Username,
			&c.User.ID,
		)
		if err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	return comments, nil
}
