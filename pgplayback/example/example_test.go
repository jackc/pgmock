package example

import (
	"context"
	"reflect"
	"testing"

	"github.com/jackc/pgmock/pgplayback"
)

var options = pgplayback.Flags()

func TestPostgres(t *testing.T) {
	ctx := context.Background()

	dbTransport := pgplayback.New("testdata/select-one-post.pgplayback", options)
	dbConn, err := dbTransport.Connect(ctx, "")
	if err != nil {
		t.Fatalf("transport failed: %v", err)
	}
	defer func() {
		if err := dbTransport.Close(ctx); err != nil {
			t.Fatalf("unexpected closing error: %v", err)
		}
	}()

	db := &Database{
		Postgres: dbConn,
	}

	req := CreatePostRequest{
		ID:    "abc",
		Title: "Fox",
		Body:  "The quick brown fox jumps over the lazy dog.",
	}

	if err := db.CreatePost(context.Background(), req); err != nil {
		t.Errorf("cannot create post: %v", err)
	}

	got, err := db.GetPost(ctx, "abc")
	if err != nil {
		t.Errorf("cannot get post: %v", err)
	}

	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if got.ModifiedAt.IsZero() {
		t.Error("ModifiedAt should not be zero")
	}

	want := &Post{
		ID:    "abc",
		Title: req.Title,
		Body:  req.Body,

		// copying already validated CreatedAt and ModifiedAt times.
		CreatedAt:  got.CreatedAt,
		ModifiedAt: got.ModifiedAt,
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("wanted %v, got %v instead", want, got)
	}

	if err := db.DeletePost(context.Background(), "abc"); err != nil {
		t.Errorf("error deleting post: %v", err)
	}
}
