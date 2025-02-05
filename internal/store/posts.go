package store

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"social/internal/types"

	"github.com/lib/pq"
)

type IPostStore interface {
	Create(ctx context.Context, post *types.Post) error
	GetPostById(ctx context.Context, postId int64) (*types.Post, error)
	Update(ctx context.Context, post *types.Post) error
	Delete(ctx context.Context, postId int64) error
	GetUserFeed(ctx context.Context, userId int64, fq PaginatedFeedQuery) ([]types.PostWithMetadata, error)
}

type PostStore struct {
	db *sql.DB
}

func (store *PostStore) Create(ctx context.Context, post *types.Post) error {

	query := `
	INSERT INTO posts (content,title,user_id,tags) 
	VALUES($1,$2,$3,$4) RETURNING id,created_at,updated_at,version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := store.db.QueryRowContext(ctx,
		query,
		post.Content,
		post.Title,
		post.UserId,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)

	if err != nil {
		return err
	}

	return nil
}

func (store *PostStore) GetPostById(ctx context.Context, postId int64) (*types.Post, error) {

	query := `
	SELECT id,user_id,title,content,created_at,updated_at,tags,version
	FROM posts
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	var post types.Post

	err := store.db.QueryRowContext(ctx, query, postId).Scan(
		&post.ID,
		&post.UserId,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (store *PostStore) Update(ctx context.Context, post *types.Post) error {

	query := `
	UPDATE posts
	SET content = $2,title = $3,version = version + 1
	WHERE id = $1 AND version = $4
	RETURNING version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := store.db.QueryRowContext(ctx,
		query,
		post.ID,
		post.Content,
		post.Title,
		post.Version,
	).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err

		}
	}

	return nil
}

func (store *PostStore) Delete(ctx context.Context, postId int64) error {

	query := `
	DELETE FROM posts WHERE id=$1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	response, err := store.db.ExecContext(ctx, query, postId)
	if err != nil {
		return err
	}

	rows, err := response.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (store *PostStore) GetUserFeed(ctx context.Context, userId int64, PaginatedFeedQuery PaginatedFeedQuery) ([]types.PostWithMetadata, error) {

	log.Printf("come from req limit: %v, offset: %v, sort: %v,search: %v", PaginatedFeedQuery.Limit, PaginatedFeedQuery.Offset, PaginatedFeedQuery.Sort, PaginatedFeedQuery.Search)

	query := `
	SELECT p.id,p.user_id,u.username,p.title,p.content,p.created_at,p.version,p.tags,
	COUNT(c.id) AS comments_count
	FROM posts p
	LEFT JOIN comments c ON c.post_id = p.id
	LEFT JOIN users u ON p.user_id = u.id
	JOIN followers f ON f.user_id = p.user_id 
	WHERE 
		f.follower_id = $1 AND 
		(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
	GROUP BY p.id, u.username
	ORDER BY p.created_at ` + PaginatedFeedQuery.Sort + `
	LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := store.db.QueryContext(ctx, query,
		userId,
		PaginatedFeedQuery.Limit,
		PaginatedFeedQuery.Offset,
		PaginatedFeedQuery.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []types.PostWithMetadata

	for rows.Next() {
		var p types.PostWithMetadata

		err := rows.Scan(
			&p.ID,
			&p.UserId,
			&p.User.Username,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.CommentsCount,
		)
		if err != nil {
			return nil, err
		}
		feed = append(feed, p)

	}

	return feed, nil

}
