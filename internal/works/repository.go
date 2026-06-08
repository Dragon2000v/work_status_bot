package works

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrNoActiveWork = errors.New("no active work")

type Repository interface {
	CreateActive(ctx context.Context, record WorkRecord) (WorkRecord, error)
	FindActiveByPerson(ctx context.Context, personID primitive.ObjectID) (WorkRecord, error)
	StopActive(ctx context.Context, personID primitive.ObjectID, stoppedAt time.Time, reason string) (WorkRecord, error)
	ListStatus(ctx context.Context) ([]WorkRecord, error)
	ListOverlappingMonth(ctx context.Context, startUTC, endUTC time.Time) ([]WorkRecord, error)
}

type WorkRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *WorkRepository {
	return &WorkRepository{collection: db.Collection("work_records")}
}

func (r *WorkRepository) CreateActive(ctx context.Context, record WorkRecord) (WorkRecord, error) {
	_, err := r.collection.InsertOne(ctx, record)
	return record, err
}

func (r *WorkRepository) FindActiveByPerson(ctx context.Context, personID primitive.ObjectID) (WorkRecord, error) {
	var record WorkRecord
	err := r.collection.FindOne(ctx, bson.M{"person_id": personID, "status": StatusActive}).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return WorkRecord{}, ErrNoActiveWork
	}
	return record, err
}

func (r *WorkRepository) StopActive(ctx context.Context, personID primitive.ObjectID, stoppedAt time.Time, reason string) (WorkRecord, error) {
	update := bson.M{"$set": bson.M{"status": StatusStopped, "stopped_at": stoppedAt.UTC(), "stop_reason": reason, "updated_at": stoppedAt.UTC()}}
	var record WorkRecord
	err := r.collection.FindOneAndUpdate(ctx, bson.M{"person_id": personID, "status": StatusActive}, update).Decode(&record)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return WorkRecord{}, ErrNoActiveWork
	}
	if err != nil {
		return WorkRecord{}, err
	}
	record.Status = StatusStopped
	record.StoppedAt = &stoppedAt
	record.StopReason = reason
	record.UpdatedAt = stoppedAt
	return record, nil
}

func (r *WorkRepository) ListStatus(ctx context.Context) ([]WorkRecord, error) {
	return r.find(ctx, bson.M{})
}

func (r *WorkRepository) ListOverlappingMonth(ctx context.Context, startUTC, endUTC time.Time) ([]WorkRecord, error) {
	filter := bson.M{
		"started_at": bson.M{"$lt": endUTC.UTC()},
		"$or": []bson.M{
			{"stopped_at": bson.M{"$exists": false}},
			{"stopped_at": nil},
			{"stopped_at": bson.M{"$gte": startUTC.UTC()}},
		},
	}
	return r.find(ctx, filter)
}

func (r *WorkRepository) find(ctx context.Context, filter bson.M) ([]WorkRecord, error) {
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []WorkRecord
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
