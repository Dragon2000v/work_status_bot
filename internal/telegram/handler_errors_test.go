package telegram

import (
	"context"
	"strings"
	"testing"
	"time"

	"work-status-bot/internal/people"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestHandlerCommandErrors(t *testing.T) {
	personRepo := newTelegramPeopleRepo()
	workRepo := &telegramWorkRepo{}
	handler := NewHandler(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, time.Now().UTC()}, nil, nil)
	for _, text := range []string{"/add_person Ivan", "/start_work Ivan Petrenko", "/bogus", "/stop_work Ivan Petrenko"} {
		got, err := handler.HandleText(context.Background(), text)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if !strings.HasPrefix(got, "Помилка:") {
			t.Fatalf("want error for %s, got %q", text, got)
		}
	}
	_, _ = personRepo.Create(context.Background(), people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")})
	if got, _ := handler.HandleText(context.Background(), "/stop_work Ivan Petrenko"); !strings.Contains(got, "no active") {
		t.Fatalf("want no active work, got %q", got)
	}
	if _, err := handler.HandleText(context.Background(), "/start_work Ivan Petrenko \"API\""); err != nil {
		t.Fatal(err)
	}
	got, _ := handler.HandleText(context.Background(), "/start_work Ivan Petrenko \"Other\"")
	if !strings.Contains(got, "active") {
		t.Fatalf("want already active, got %q", got)
	}
}
