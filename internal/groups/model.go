package groups

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const UnknownGroupTitle = "Unknown group"

const (
	ReasonConfigured       = "configured"
	ReasonFallback         = "fallback"
	ReasonSetupAllowed     = "setup_allowed"
	ReasonHelpAllowed      = "help_allowed"
	ReasonUnknownGroup     = "unknown_group"
	ReasonDisabledGroup    = "disabled_group"
	ReasonPrivateChatSetup = "private_chat_setup"
	ReasonPrivateChat      = "private_chat"
)

type ConfiguredGroup struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TelegramChatID int64              `bson:"telegram_chat_id" json:"telegram_chat_id"`
	Title          string             `bson:"title" json:"title"`
	SetupUserID    int64              `bson:"setup_user_id" json:"setup_user_id"`
	SetupUsername  string             `bson:"setup_username,omitempty" json:"setup_username,omitempty"`
	Enabled        bool               `bson:"enabled" json:"enabled"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

type GroupSetupRequest struct {
	TelegramChatID int64
	Title          string
	SetupUserID    int64
	SetupUsername  string
	OccurredAt     time.Time
}

type SetupResult struct {
	Group   ConfiguredGroup
	Outcome string
}

type GroupAuthorizationDecision struct {
	ChatID  int64
	Command string
	Allowed bool
	Reason  string
}

type ReportDeliveryTarget struct {
	TelegramChatID int64
	Source         string
	Title          string
}

const (
	ReportTargetCurrentChat     = "current_chat"
	ReportTargetConfiguredGroup = "configured_group"
	ReportTargetFallbackGroup   = "fallback_group"
)
