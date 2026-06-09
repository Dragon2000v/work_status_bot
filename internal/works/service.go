package works

import (
	"context"
	"errors"
	"strings"
	"time"

	"work-status-bot/internal/people"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrMissingTitle      = errors.New("work title is required")
	ErrAlreadyActiveWork = errors.New("person already has active work")
)

type Service struct {
	people people.Repository
	works  Repository
	now    func() time.Time
}

func NewService(peopleRepo people.Repository, workRepo Repository) *Service {
	return &Service{people: peopleRepo, works: workRepo, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) SetNow(now func() time.Time) {
	s.now = now
}

func (s *Service) Start(ctx context.Context, firstName, lastName, title string) (people.Person, WorkRecord, error) {
	return s.StartAt(ctx, firstName, lastName, title, s.now())
}

func (s *Service) StartAt(ctx context.Context, firstName, lastName, title string, startedAt time.Time) (people.Person, WorkRecord, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return people.Person{}, WorkRecord{}, ErrMissingTitle
	}
	person, err := s.people.FindByNormalizedName(ctx, people.NormalizeFullName(firstName, lastName))
	if err != nil {
		return people.Person{}, WorkRecord{}, err
	}
	if _, err := s.works.FindActiveByPerson(ctx, person.ID); err == nil {
		return people.Person{}, WorkRecord{}, ErrAlreadyActiveWork
	} else if !errors.Is(err, ErrNoActiveWork) {
		return people.Person{}, WorkRecord{}, err
	}
	now := s.now().UTC()
	startedAt = startedAt.UTC()
	record := WorkRecord{
		ID: primitive.NewObjectID(), PersonID: person.ID, Title: title, Status: StatusActive,
		StartedAt: startedAt, CreatedAt: now, UpdatedAt: now,
	}
	record, err = s.works.CreateActive(ctx, record)
	return person, record, err
}

func (s *Service) Status(ctx context.Context) ([]people.Person, []WorkRecord, error) {
	persons, err := s.people.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	records, err := s.works.ListStatus(ctx)
	if err != nil {
		return nil, nil, err
	}
	return persons, records, nil
}

func (s *Service) Stop(ctx context.Context, firstName, lastName, reason string) (people.Person, WorkRecord, error) {
	person, err := s.people.FindByNormalizedName(ctx, people.NormalizeFullName(firstName, lastName))
	if err != nil {
		return people.Person{}, WorkRecord{}, err
	}
	record, err := s.works.StopActive(ctx, person.ID, s.now().UTC(), strings.TrimSpace(reason))
	return person, record, err
}

func (s *Service) ListOverlappingMonth(ctx context.Context, month string) ([]WorkRecord, error) {
	start, end, err := MonthBoundsKyiv(month)
	if err != nil {
		return nil, err
	}
	return s.works.ListOverlappingMonth(ctx, start, end)
}
