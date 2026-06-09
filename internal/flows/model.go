package flows

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const Expiration = 15 * time.Minute

const (
	TypeAddPerson = "add_person"
	TypeStartWork = "start_work"
	TypeStopWork  = "stop_work"
	TypeSettings  = "settings"
)

const (
	StepAddPersonEnterName           = "add_person_enter_name"
	StepStartWorkEnterPerson         = "start_work_enter_person"
	StepStartWorkEnterTitle          = "start_work_enter_title"
	StepStartWorkSelectTimeMode      = "start_work_select_time_mode"
	StepStartWorkEnterTime           = "start_work_enter_time"
	StepStartWorkEnterManualDateTime = "start_work_enter_manual_datetime"
	StepStopWorkEnterPersonOrReason  = "stop_work_enter_person_or_reason"
	StepSettingsLanguageSelect       = "settings_language_select"
)

type State struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TelegramUserID int64              `bson:"telegram_user_id" json:"telegram_user_id"`
	ChatID         int64              `bson:"chat_id" json:"chat_id"`
	FlowType       string             `bson:"flow_type" json:"flow_type"`
	Step           string             `bson:"step" json:"step"`
	Payload        map[string]string  `bson:"payload" json:"payload"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
	ExpiresAt      time.Time          `bson:"expires_at" json:"expires_at"`
}

func NewState(userID, chatID int64, flowType, step string, payload map[string]string, now time.Time) State {
	now = now.UTC()
	return State{
		ID: primitive.NewObjectID(), TelegramUserID: userID, ChatID: chatID,
		FlowType: flowType, Step: step, Payload: clonePayload(payload),
		CreatedAt: now, UpdatedAt: now, ExpiresAt: now.Add(Expiration),
	}
}

func (s State) Expired(now time.Time) bool {
	return !now.UTC().Before(s.ExpiresAt)
}

func (s State) WithStep(step string, payload map[string]string, now time.Time) State {
	now = now.UTC()
	s.Step = step
	s.Payload = clonePayload(payload)
	s.UpdatedAt = now
	s.ExpiresAt = now.Add(Expiration)
	return s
}

func clonePayload(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
