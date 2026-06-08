package users

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("user setting not found")

type Repository interface {
	GetByTelegramUserID(ctx context.Context, telegramUserID int64) (UserSetting, error)
	UpsertLanguage(ctx context.Context, telegramUserID int64, language string, now time.Time) (UserSetting, error)
}

type MongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{collection: db.Collection("user_settings")}
}

func (r *MongoRepository) GetByTelegramUserID(ctx context.Context, telegramUserID int64) (UserSetting, error) {
	var setting UserSetting
	err := r.collection.FindOne(ctx, bson.M{"telegram_user_id": telegramUserID}).Decode(&setting)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return UserSetting{}, ErrNotFound
	}
	return setting, err
}

func (r *MongoRepository) UpsertLanguage(ctx context.Context, telegramUserID int64, language string, now time.Time) (UserSetting, error) {
	now = now.UTC()
	filter := bson.M{"telegram_user_id": telegramUserID}
	update := bson.M{
		"$set": bson.M{
			"language":   language,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"telegram_user_id": telegramUserID,
			"created_at":       now,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var setting UserSetting
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&setting); err != nil {
		return UserSetting{}, err
	}
	return setting, nil
}
