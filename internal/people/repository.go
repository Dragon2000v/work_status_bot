package people

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrNotFound = errors.New("person not found")

type Repository interface {
	Create(ctx context.Context, person Person) (Person, error)
	FindByNormalizedName(ctx context.Context, normalized string) (Person, error)
	List(ctx context.Context) ([]Person, error)
}

type PeopleRepository struct {
	collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *PeopleRepository {
	return &PeopleRepository{collection: db.Collection("people")}
}

func (r *PeopleRepository) Create(ctx context.Context, person Person) (Person, error) {
	res, err := r.collection.InsertOne(ctx, person)
	if err != nil {
		return Person{}, err
	}
	if id, ok := res.InsertedID.(interface{ Hex() string }); ok && id.Hex() != "" {
		// ObjectID is already encoded by Mongo when omitted; caller does not need it for command response.
	}
	return person, nil
}

func (r *PeopleRepository) FindByNormalizedName(ctx context.Context, normalized string) (Person, error) {
	var person Person
	err := r.collection.FindOne(ctx, bson.M{"normalized_full_name": normalized}).Decode(&person)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Person{}, ErrNotFound
	}
	return person, err
}

func (r *PeopleRepository) List(ctx context.Context) ([]Person, error) {
	cur, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []Person
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}
