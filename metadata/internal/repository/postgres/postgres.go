package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"movieexample.com/metadata/internal/repository"
	"movieexample.com/metadata/pkg/model"
)

const (
	host    = "postgres"
	port    = 5432
	user    = "postgres"
	password = "Password@123"
	dbname  = "movieapp"
	sslmode = "disable"
)

// Repository defines a Postgres-based movie metadata repository.
type Repository struct {
	db *sql.DB
}

// New creates a new Postgres-based repository.
func New() (*Repository, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return &Repository{db}, nil
}

// Get retrieves movie metadata by movie id
func (r *Repository) Get(ctx context.Context, id string) (*model.Metadata, error) {
	var title, description, director string
	row := r.db.QueryRowContext(ctx, "SELECT title, description, director FROM movies WHERE id = $1", id)
	if err := row.Scan(&title, &description, &director); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	
	return &model.Metadata{
		ID:				id,
		Title:			title,
		Description:	description,
		Director:		director,
	}, nil
}

// Put adds movie metadata for a given movie id.
func (r *Repository) Put(ctx context.Context, id string, metadata *model.Metadata) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO movies(id, title, description, director) VALUES ($1, $2, $3, $4)", id, metadata.Title, metadata.Description, metadata.Director)
	return err
}