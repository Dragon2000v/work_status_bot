package works

import (
	"context"
	"errors"
	"testing"
	"time"

	"work-status-bot/internal/people"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type memWorkRepo struct {
	records []WorkRecord
}

type workPeopleRepo struct {
	people map[string]people.Person
}

func newWorkPeopleRepo() *workPeopleRepo {
	return &workPeopleRepo{people: map[string]people.Person{}}
}

func (r *workPeopleRepo) Create(ctx context.Context, person people.Person) (people.Person, error) {
	r.people[person.NormalizedFullName] = person
	return person, nil
}

func (r *workPeopleRepo) FindByNormalizedName(ctx context.Context, normalized string) (people.Person, error) {
	p, ok := r.people[normalized]
	if !ok {
		return people.Person{}, people.ErrNotFound
	}
	return p, nil
}

func (r *workPeopleRepo) List(ctx context.Context) ([]people.Person, error) {
	var out []people.Person
	for _, p := range r.people {
		out = append(out, p)
	}
	return out, nil
}

func (r *memWorkRepo) CreateActive(ctx context.Context, record WorkRecord) (WorkRecord, error) {
	r.records = append(r.records, record)
	return record, nil
}

func (r *memWorkRepo) FindActiveByPerson(ctx context.Context, personID primitive.ObjectID) (WorkRecord, error) {
	for _, rec := range r.records {
		if rec.PersonID == personID && rec.Status == StatusActive {
			return rec, nil
		}
	}
	return WorkRecord{}, ErrNoActiveWork
}

func (r *memWorkRepo) StopActive(ctx context.Context, personID primitive.ObjectID, stoppedAt time.Time, reason string) (WorkRecord, error) {
	for i := range r.records {
		if r.records[i].PersonID == personID && r.records[i].Status == StatusActive {
			r.records[i].Status = StatusStopped
			r.records[i].StoppedAt = &stoppedAt
			r.records[i].StopReason = reason
			return r.records[i], nil
		}
	}
	return WorkRecord{}, ErrNoActiveWork
}

func (r *memWorkRepo) ListStatus(ctx context.Context) ([]WorkRecord, error) { return r.records, nil }

func (r *memWorkRepo) ListOverlappingMonth(ctx context.Context, startUTC, endUTC time.Time) ([]WorkRecord, error) {
	var out []WorkRecord
	for _, rec := range r.records {
		stopped := endUTC
		if rec.StoppedAt != nil {
			stopped = *rec.StoppedAt
		}
		if rec.StartedAt.Before(endUTC) && !stopped.Before(startUTC) {
			out = append(out, rec)
		}
	}
	return out, nil
}

func TestWorkServiceStartStatusStop(t *testing.T) {
	personRepo := newWorkPeopleRepo()
	p, _ := personRepo.Create(context.Background(), people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")})
	workRepo := &memWorkRepo{}
	svc := NewService(personRepo, workRepo)
	start := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return start }
	_, rec, err := svc.Start(context.Background(), p.FirstName, p.LastName, "API fix")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, _, err := svc.Start(context.Background(), p.FirstName, p.LastName, "Other"); !errors.Is(err, ErrAlreadyActiveWork) {
		t.Fatalf("want active error, got %v", err)
	}
	svc.now = func() time.Time { return start.Add(90 * time.Minute) }
	if Elapsed(rec, svc.now()) != 90*time.Minute {
		t.Fatalf("elapsed from stored timestamps wrong")
	}
	_, stopped, err := svc.Stop(context.Background(), p.FirstName, p.LastName, "done")
	if err != nil {
		t.Fatalf("stop: %v", err)
	}
	if stopped.Status != StatusStopped || stopped.StoppedAt == nil {
		t.Fatalf("record not stopped")
	}
	if _, _, err := svc.Stop(context.Background(), p.FirstName, p.LastName, "again"); !errors.Is(err, ErrNoActiveWork) {
		t.Fatalf("want no active, got %v", err)
	}
}
