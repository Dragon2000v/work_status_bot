package people

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Person struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName          string             `bson:"first_name" json:"first_name"`
	LastName           string             `bson:"last_name" json:"last_name"`
	NormalizedFullName string             `bson:"normalized_full_name" json:"normalized_full_name"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

func NormalizeFullName(firstName, lastName string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(firstName)+" "+strings.TrimSpace(lastName)), " "))
}
