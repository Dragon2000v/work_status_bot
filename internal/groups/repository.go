package groups

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("configured group not found")

type Repository interface {
	UpsertSetup(ctx context.Context, req GroupSetupRequest, now time.Time) (ConfiguredGroup, bool, bool, error)
	FindByTelegramChatID(ctx context.Context, chatID int64) (ConfiguredGroup, error)
	ListEnabled(ctx context.Context) ([]ConfiguredGroup, error)
	ListAll(ctx context.Context) ([]ConfiguredGroup, error)
	Disable(ctx context.Context, chatID int64, now time.Time) (ConfiguredGroup, bool, error)
}

type MongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{collection: db.Collection("configured_groups")}
}

func (r *MongoRepository) UpsertSetup(ctx context.Context, req GroupSetupRequest, now time.Time) (ConfiguredGroup, bool, bool, error) {
	now = now.UTC()
	title := NormalizeTitle(req.Title)
	filter := bson.M{"telegram_chat_id": req.TelegramChatID}
	update := bson.M{
		"$set": bson.M{
			"title":          title,
			"setup_user_id":  req.SetupUserID,
			"setup_username": req.SetupUsername,
			"enabled":        true,
			"updated_at":     now,
		},
		"$setOnInsert": bson.M{
			"telegram_chat_id": req.TelegramChatID,
			"created_at":       now,
		},
	}
	var before ConfiguredGroup
	err := r.collection.FindOne(ctx, filter).Decode(&before)
	existed := err == nil
	wasDisabled := existed && !before.Enabled
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return ConfiguredGroup{}, false, false, err
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var group ConfiguredGroup
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&group); err != nil {
		return ConfiguredGroup{}, false, false, err
	}
	return group, existed, wasDisabled, nil
}

func (r *MongoRepository) FindByTelegramChatID(ctx context.Context, chatID int64) (ConfiguredGroup, error) {
	var group ConfiguredGroup
	err := r.collection.FindOne(ctx, bson.M{"telegram_chat_id": chatID}).Decode(&group)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ConfiguredGroup{}, ErrNotFound
	}
	return group, err
}

func (r *MongoRepository) ListEnabled(ctx context.Context) ([]ConfiguredGroup, error) {
	return r.list(ctx, bson.M{"enabled": true})
}

func (r *MongoRepository) ListAll(ctx context.Context) ([]ConfiguredGroup, error) {
	return r.list(ctx, bson.M{})
}

func (r *MongoRepository) list(ctx context.Context, filter bson.M) ([]ConfiguredGroup, error) {
	cur, err := r.collection.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "title", Value: 1}, {Key: "telegram_chat_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []ConfiguredGroup
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *MongoRepository) Disable(ctx context.Context, chatID int64, now time.Time) (ConfiguredGroup, bool, error) {
	now = now.UTC()
	before, err := r.FindByTelegramChatID(ctx, chatID)
	if err != nil {
		return ConfiguredGroup{}, false, err
	}
	alreadyDisabled := !before.Enabled
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var group ConfiguredGroup
	err = r.collection.FindOneAndUpdate(ctx, bson.M{"telegram_chat_id": chatID}, bson.M{"$set": bson.M{"enabled": false, "updated_at": now}}, opts).Decode(&group)
	if err != nil {
		return ConfiguredGroup{}, false, err
	}
	return group, alreadyDisabled, nil
}
