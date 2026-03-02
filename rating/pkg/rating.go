package model

// RecordID defines a record id. Together with RecordType
// identifies unique records across all types.
type RecordID string

// RecordType defines a record id. Together with RecordID
// identifies unique records across all types.
type RecordType string

// Existing Record Types
const (
	RecordTypeMovie = RecordType("movie")
)

// RatingValue defines a value of a rating record.
type RatingValue int

type Rating struct {
	RecordID 	string		`json:recordID`
	RecordType	string		`json:recordType`
	UserID		string		`json:userID`
	Value		RatingValue	`json:value`
}