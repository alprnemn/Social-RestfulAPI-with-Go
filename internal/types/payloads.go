package types

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

type FollowUserPayload struct {
	UserID    int64  `json:"user_id" validate:"required"`
	CreatedAt string `json:"created_at"`
}

type RegisterUserPayload struct {
	Username       string `json:"username" validate:"required,min=2,max=50"`
	Email          string `json:"email" validate:"required,email,max=255"`
	Password       string `json:"password" validate:"required,min=3,max=72"`
	HashedPassword string
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}
