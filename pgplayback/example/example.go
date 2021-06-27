package example

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgmock/pgplayback"
)

// Database DAO.
type Database struct {
	Postgres pgplayback.Postgres
}

// Post on the database.
type Post struct {
	ID         string
	Title      string
	Body       string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

// CreatePostRequest is used by CreatePost to create a post.
type CreatePostRequest struct {
	ID    string
	Title string
	Body  string
}

// CreatePost in the database.
func (db *Database) CreatePost(ctx context.Context, req CreatePostRequest) error {
	const sql = "INSERT INTO posts (id, title, body) VALUES ( $1, $2, $3 )"
	_, err := db.Postgres.Exec(ctx, sql, req.ID, req.Title, req.Body)
	return err
}

// GetPost from the database.
func (db *Database) GetPost(ctx context.Context, id string) (*Post, error) {
	var p Post
	row := db.Postgres.QueryRow(ctx, "SELECT id, title, body, created_at, modified_at FROM posts WHERE id = $1", id)
	err := row.Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.ModifiedAt)
	return &p, err
}

// DeletePost from the database.
func (db *Database) DeletePost(ctx context.Context, id string) error {
	if c, err := db.Postgres.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id); err != nil {
		return fmt.Errorf("cannot remove post: %w (%d rows affected)", err, c.RowsAffected())
	}
	return nil
}
