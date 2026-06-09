package flows

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("flow state not found")

type Repository interface {
	Upsert(ctx context.Context, state State) (State, error)
	Get(ctx context.Context, telegramUserID, chatID int64) (State, error)
	Delete(ctx context.Context, telegramUserID, chatID int64) error
}

type MongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{collection: db.Collection("user_flow_states")}
}

func (r *MongoRepository) Upsert(ctx context.Context, state State) (State, error) {
	filter := bson.M{"telegram_user_id": state.TelegramUserID, "chat_id": state.ChatID}
	update := bson.M{
		"$set": bson.M{
			"flow_type":  state.FlowType,
			"step":       state.Step,
			"payload":    state.Payload,
			"updated_at": state.UpdatedAt.UTC(),
			"expires_at": state.ExpiresAt.UTC(),
		},
		"$setOnInsert": bson.M{
			"_id":              state.ID,
			"telegram_user_id": state.TelegramUserID,
			"chat_id":          state.ChatID,
			"created_at":       state.CreatedAt.UTC(),
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var out State
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return State{}, err
	}
	return out, nil
}

func (r *MongoRepository) Get(ctx context.Context, telegramUserID, chatID int64) (State, error) {
	var state State
	err := r.collection.FindOne(ctx, bson.M{"telegram_user_id": telegramUserID, "chat_id": chatID}).Decode(&state)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return State{}, ErrNotFound
	}
	return state, err
}

func (r *MongoRepository) Delete(ctx context.Context, telegramUserID, chatID int64) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"telegram_user_id": telegramUserID, "chat_id": chatID})
	return err
}
