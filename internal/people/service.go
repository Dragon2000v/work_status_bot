package people

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrInvalidPerson   = errors.New("first and last name are required")
	ErrDuplicatePerson = errors.New("person already exists")
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Add(ctx context.Context, firstName, lastName string) (Person, error) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	if firstName == "" || lastName == "" {
		return Person{}, ErrInvalidPerson
	}
	normalized := NormalizeFullName(firstName, lastName)
	if _, err := s.repo.FindByNormalizedName(ctx, normalized); err == nil {
		return Person{}, ErrDuplicatePerson
	} else if !errors.Is(err, ErrNotFound) {
		return Person{}, err
	}
	now := s.now().UTC()
	return s.repo.Create(ctx, Person{
		ID: primitive.NewObjectID(), FirstName: firstName, LastName: lastName,
		NormalizedFullName: normalized, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *Service) Find(ctx context.Context, firstName, lastName string) (Person, error) {
	return s.repo.FindByNormalizedName(ctx, NormalizeFullName(firstName, lastName))
}

func (s *Service) List(ctx context.Context) ([]Person, error) {
	return s.repo.List(ctx)
}
