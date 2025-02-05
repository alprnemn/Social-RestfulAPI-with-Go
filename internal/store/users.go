package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"
	"social/internal/types"
	"time"

	"github.com/lib/pq"
)

type IUserStore interface {
	Create(context.Context, *sql.Tx, *types.User) error
	Delete(context.Context, int64) error
	GetByEmail(context.Context, string) (*types.User, error)
	GetById(context.Context, int64) (*types.User, error)
	Follow(context.Context, int64, int64) error
	Unfollow(context.Context, int64, int64) error
	RegisterUser(context.Context, *types.RegisterUserPayload) error
	CreateAndInvite(context.Context, *types.User, string, time.Duration) error
	Activate(context.Context, string) error
}

type UserStore struct {
	db *sql.DB
}

func (store *UserStore) Create(ctx context.Context, tx *sql.Tx, user *types.User) error {

	query := `
	INSERT INTO users (username,email,password,role_id)
	VALUES ($1,$2,$3,(SELECT id FROM roles WHERE name = $4)) 
	RETURNING id,created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	role := user.Role.Name
	if role == "" {
		role = "user"
	}

	err := tx.QueryRowContext(ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
		user.RoleId,
		role,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err != nil {
		switch err.Error() {
		case `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		case `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return ErrWDFUQ
		}
	}

	return nil
}

func (store *UserStore) GetByEmail(ctx context.Context, email string) (*types.User, error) {

	query := `SELECT id,username,email,password,created_at,is_active FROM users WHERE email=$1 AND is_active = true`

	// we should check user is active every time

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &types.User{}

	err := store.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.IsActive,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil

}

func (store *UserStore) GetById(ctx context.Context, userId int64) (*types.User, error) {

	query := `
	SELECT users.id,username,email,password,created_at,roles.*
	FROM users
	JOIN rules ON (users.role_id = roles.id)
	WHERE users.id = $1 AND is_active = true
	`

	ctx, close := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer close()

	user := new(types.User)

	err := store.db.QueryRowContext(ctx, query, userId).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (store *UserStore) Follow(ctx context.Context, userId int64, followerId int64) error {

	query := `
	INSERT INTO followers (user_id,follower_id) VALUES ($1,$2)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := store.db.ExecContext(ctx, query, userId, followerId)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
	}

	return nil
}

func (store *UserStore) Unfollow(ctx context.Context, userId int64, unfollowerId int64) error {

	query := `
	DELETE FROM followers
	WHERE user_id = $1 AND	follower_id= $2
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := store.db.ExecContext(ctx, query, userId, unfollowerId)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
	}

	return nil
}

func (store *UserStore) RegisterUser(ctx context.Context, user *types.RegisterUserPayload) error {

	query := `
	INSERT INTO users(username,email,password)
	VALUES ($1,$2,$3)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	if _, err := store.db.ExecContext(ctx, query, user.Username, user.Email, user.Password); err != nil {

		switch err.Error() {
		case `pq: duplicate key value violates unique constraint "users_username_key"`:
			return ErrDuplicateUsername
		case `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return ErrWDFUQ
		}
	}

	return nil
}

func (store *UserStore) CreateAndInvite(ctx context.Context, user *types.User, token string, invitationExp time.Duration) error {

	// transaction wrapper
	return withTx(store.db, ctx, func(tx *sql.Tx) error {

		// create the user
		if err := store.Create(ctx, tx, user); err != nil {
			return err
		}

		// create the user invite
		err := store.createUserInvitation(ctx, tx, token, invitationExp, user.ID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (store *UserStore) Activate(ctx context.Context, token string) error {

	return withTx(store.db, ctx, func(tx *sql.Tx) error {

		// find the user that this token belongs to
		user, err := store.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			return err
		}

		// update the user
		user.IsActive = true
		if err := store.updateUser(ctx, tx, user); err != nil {
			return err
		}

		// clean the invitations
		if err := store.deleteUserInvitations(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (store *UserStore) updateUser(ctx context.Context, tx *sql.Tx, user *types.User) error {

	query := `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.Username, user.Email, user.IsActive, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*types.User, error) {
	query := `
	SELECT u.id,u.username,u.email,u.created_at,u.is_active
	FROM users u
	JOIN user_invitations ui ON u.id = ui.user_id
	WHERE ui.token = $1 AND ui.expiry > $2
	`

	hash := sha256.Sum256([]byte(token))
	hashToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &types.User{}
	err := tx.QueryRowContext(ctx, query, hashToken, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActive,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			return nil, ErrWDFUQ
		}
	}

	return user, nil
}

func (store *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, invitationExp time.Duration, userId int64) error {

	log.Println("token: ", token)

	query := `INSERT INTO user_invitations (token,user_id,expiry) VALUES ($1,$2,$3)`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, token, userId, time.Now().Add(invitationExp))
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) delete(ctx context.Context, tx *sql.Tx, userID int64) error {

	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) Delete(ctx context.Context, userID int64) error {
	return withTx(store.db, ctx, func(tx *sql.Tx) error {

		if err := store.delete(ctx, tx, userID); err != nil {
			return err
		}

		if err := store.deleteUserInvitations(ctx, tx, userID); err != nil {
			return err
		}

		return nil

	})
}
