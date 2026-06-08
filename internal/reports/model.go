package reports

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	IncidentStopped     = "stopped"
	IncidentAlertFailed = "alert_failed"

	TriggerTelegramCommand = "telegram_command"
	TriggerExternalHTTP    = "external_http"
)

type Incident struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PersonID     primitive.ObjectID `bson:"person_id" json:"person_id"`
	WorkRecordID primitive.ObjectID `bson:"work_record_id" json:"work_record_id"`
	Type         string             `bson:"type" json:"type"`
	Reason       string             `bson:"reason,omitempty" json:"reason,omitempty"`
	OccurredAt   time.Time          `bson:"occurred_at" json:"occurred_at"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

type MonthlyReport struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Month         string             `bson:"month" json:"month"`
	TriggerSource string             `bson:"trigger_source" json:"trigger_source"`
	Content       string             `bson:"content" json:"content"`
	GeneratedAt   time.Time          `bson:"generated_at" json:"generated_at"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}
