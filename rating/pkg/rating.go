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

// UserID defines a user id.
type UserID string

// RatingValue defines a value of a rating record.
type RatingValue int

type Rating struct {
	RecordID   string      `json:"recordID"`
	RecordType string      `json:"recordType"`
	UserID     UserID      `json:"userId"`
	Value      RatingValue `json:"value"`
}

// RatingEvent defines an event containing rating information.
type RatingEvent struct {
	Rating
	ProviderID string          `json:"providerId"`
	EventType  RatingEventType `json:"eventType"`
}

// RatingEventType defines a type of a rating event.
type RatingEventType string

// Rating event types.
const (
	RatingEventTypePut    = RatingEventType("put")
	RatingEventTypeDelete = RatingEventType("delete")
)
