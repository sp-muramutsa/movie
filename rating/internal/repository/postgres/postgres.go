package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"movieexample.com/rating/pkg"
)

const (
	host = "postgres"
	port = 5432
	user = "postgres"
	password = "Password@123"
	dbname = "movieapp"
	sslmode = "disable"
)

// Repository defines a Postgres-based rating repository.
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

// Get retrieves all ratings for a given record.
func (r *Repository) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT user_id, value FROM ratings WHERE record_id = $1 AND record_type = $2",
		recordID,
		recordType,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var res []model.Rating
	for rows.Next() {
		var userID string
		var value int32
		if err := rows.Scan(&userID, &value); err != nil {
			return nil, err
		}
		res = append(res, model.Rating{
			UserID: model.UserID(userID),
			Value: model.RatingValue(value),
		})
	}

	return res, nil
}

// Put adds a rating for a give record.
func (r *Repository) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	if rating == nil {
		return errors.New("empty rating")
	}

	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO ratings (record_id, record_type, user_id, value) VALUES($1, $2, $3, $4)",
		recordID,
		recordType,
		rating.UserID,
		rating.Value,
	)
	return err
}