package users

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserSetting struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	TelegramUserID int64              `bson:"telegram_user_id" json:"telegram_user_id"`
	Language       string             `bson:"language" json:"language"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
