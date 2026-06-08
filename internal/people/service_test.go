package people

import (
	"context"
	"errors"
	"testing"
)

type memPeopleRepo struct {
	people map[string]Person
}

func newMemPeopleRepo() *memPeopleRepo {
	return &memPeopleRepo{people: map[string]Person{}}
}

func (r *memPeopleRepo) Create(ctx context.Context, person Person) (Person, error) {
	r.people[person.NormalizedFullName] = person
	return person, nil
}

func (r *memPeopleRepo) FindByNormalizedName(ctx context.Context, normalized string) (Person, error) {
	p, ok := r.people[normalized]
	if !ok {
		return Person{}, ErrNotFound
	}
	return p, nil
}

func (r *memPeopleRepo) List(ctx context.Context) ([]Person, error) {
	var out []Person
	for _, p := range r.people {
		out = append(out, p)
	}
	return out, nil
}

func TestPeopleServiceAddAndDuplicate(t *testing.T) {
	repo := newMemPeopleRepo()
	svc := NewService(repo)
	if _, err := svc.Add(context.Background(), "Ivan", "Petrenko"); err != nil {
		t.Fatalf("add person: %v", err)
	}
	if _, err := svc.Add(context.Background(), " ivan ", "PETRENKO"); !errors.Is(err, ErrDuplicatePerson) {
		t.Fatalf("want duplicate, got %v", err)
	}
}
