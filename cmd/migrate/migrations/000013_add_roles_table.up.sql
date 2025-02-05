CREATE TABLE IF NOT EXISTS roles(
    id bigserial PRIMARY KEY,
    name VARCHAR(25) UNIQUE NOT NULL,
    level int NOT NULL DEFAULT 0,
    description VARCHAR(255)
);

INSERT INTO roles (name, level, description)
VALUES
    ('user', 1, 'A user can create posts and comments'),
    ('moderator', 2, 'A moderator can update other users posts'),
    ('admin', 3, 'A admin can do everything');