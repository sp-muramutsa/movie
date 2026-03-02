package model

import (
	"movieexample.com/metadata/model"
	model "movieexample.com/metadata/pkg"
)

// MovieDetails includes movie metadata and its aggregated rating.
type MovieDetails struct {
	Rating 		*float64		`json:"rating,omitEmpty"`
	Metadata	model.Metadata	`json:"metadata"`
}