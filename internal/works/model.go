package works

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusActive  = "active"
	StatusStopped = "stopped"
)

type WorkRecord struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PersonID   primitive.ObjectID `bson:"person_id" json:"person_id"`
	Title      string             `bson:"title" json:"title"`
	Status     string             `bson:"status" json:"status"`
	StartedAt  time.Time          `bson:"started_at" json:"started_at"`
	StoppedAt  *time.Time         `bson:"stopped_at,omitempty" json:"stopped_at,omitempty"`
	StopReason string             `bson:"stop_reason,omitempty" json:"stop_reason,omitempty"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
