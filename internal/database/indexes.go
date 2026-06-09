package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	if db == nil {
		return nil
	}
	if _, err := db.Collection("people").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "normalized_full_name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}
	if _, err := db.Collection("work_records").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "person_id", Value: 1}}},
		{Keys: bson.D{{Key: "person_id", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "started_at", Value: 1}}},
		{Keys: bson.D{{Key: "stopped_at", Value: 1}}},
	}); err != nil {
		return err
	}
	if _, err := db.Collection("incidents").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "occurred_at", Value: 1}}}); err != nil {
		return err
	}
	if _, err := db.Collection("monthly_reports").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "month", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}
	if _, err := db.Collection("user_settings").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "telegram_user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}
	if _, err := db.Collection("configured_groups").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "telegram_chat_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{Keys: bson.D{{Key: "enabled", Value: 1}}},
	}); err != nil {
		return err
	}
	_, err := db.Collection("telegram_group_config").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "chat_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}
