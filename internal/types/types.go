package types

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	Tags      []string  `json:"tags"`
	UserId    int64     `json:"user_id"`              // id bigint
	CreatedAt string    `json:"created_at"`           // created_at timestamp (0)
	UpdatedAt string    `json:"updated_at,omitempty"` // updated_at timestamp (0)
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
	Version   int
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
}

type User struct {
	ID        int64  `json:"id,omitempty"`         // id bigint
	Username  string `json:"username,omitempty"`   // username varchar(255)
	Email     string `json:"email,omitempty"`      // email citext
	Password  string `json:"-"`                    // password bytea
	CreatedAt string `json:"created_at,omitempty"` // created_at timestamp (0)
	IsActive  bool   `json:"is_active,omitempty"`
	RoleId    int64  `json:"role_id"`
	Role      Role   `json:"role"`
}

type UserWithToken struct {
	*User
	Token string `json:"token"`
}

// type password struct {
// 	text *string
// 	hash []byte
// }

// func (p *password) Set(text string) error {
// 	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
// 	if err != nil {
// 		return err
// 	}
// 	p.text = &text
// 	p.hash = hash
// 	return nil
// }

type Comment struct {
	ID        int64  `json:"id"`         // id bigint
	PostId    int64  `json:"post_id"`    // post_id bigint
	UserId    int64  `json:"user_id"`    // user_id bigint
	Content   string `json:"content"`    // content text
	CreatedAt string `json:"created_at"` // created_at timestamp(0)
	User      User   `json:"user"`
}

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Level       int64  `json:"level"`
	Description string `json:"description"`
}
