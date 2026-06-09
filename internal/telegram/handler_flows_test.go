package telegram

import (
	"context"
	"strings"
	"testing"
	"time"

	"work-status-bot/internal/flows"
	"work-status-bot/internal/people"
	"work-status-bot/internal/works"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type telegramFlowRepo struct {
	state flows.State
	ok    bool
}

func (r *telegramFlowRepo) Upsert(ctx context.Context, state flows.State) (flows.State, error) {
	r.state = state
	r.ok = true
	return state, nil
}

func (r *telegramFlowRepo) Get(ctx context.Context, userID, chatID int64) (flows.State, error) {
	if !r.ok {
		return flows.State{}, flows.ErrNotFound
	}
	return r.state, nil
}

func (r *telegramFlowRepo) Delete(ctx context.Context, userID, chatID int64) error {
	r.ok = false
	return nil
}

func newFlowHandler(now time.Time) (*Handler, *telegramPeopleRepo, *telegramWorkRepo, *telegramFlowRepo) {
	personRepo := newTelegramPeopleRepo()
	workRepo := &telegramWorkRepo{}
	flowRepo := &telegramFlowRepo{}
	flowSvc := flows.NewService(flowRepo)
	flowSvc.SetNow(func() time.Time { return now })
	handler := NewHandler(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, now}, nil, nil)
	handler.SetFlows(flowSvc)
	handler.now = func() time.Time { return now }
	return handler, personRepo, workRepo, flowRepo
}

func TestAddPersonButtonFlowCreatesPersonAndClearsState(t *testing.T) {
	handler, personRepo, _, flowRepo := newFlowHandler(time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC))
	got, err := handler.HandleTextForChat(context.Background(), "Додати людину", 77, "", Chat{ID: 10, Type: "group"})
	if err != nil {
		t.Fatal(err)
	}
	if flowRepo.state.Step != flows.StepAddPersonEnterName || !strings.Contains(got, "Введіть") {
		t.Fatalf("flow not started: %q %#v", got, flowRepo.state)
	}
	got, err = handler.HandleTextForChat(context.Background(), "Іван Петренко", 77, "", Chat{ID: 10, Type: "group"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Додано") || flowRepo.ok {
		t.Fatalf("flow not completed: %q ok=%v", got, flowRepo.ok)
	}
	if _, err := personRepo.FindByNormalizedName(context.Background(), people.NormalizeFullName("Іван", "Петренко")); err != nil {
		t.Fatalf("person not created: %v", err)
	}
}

func TestStopWorkButtonFlowStopsActiveWork(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	handler, personRepo, workRepo, flowRepo := newFlowHandler(now)
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Іван", LastName: "Петренко", NormalizedFullName: people.NormalizeFullName("Іван", "Петренко")}
	_, _ = personRepo.Create(context.Background(), p)
	workRepo.records = append(workRepo.records, works.WorkRecord{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "API", Status: works.StatusActive, StartedAt: now})
	if _, err := handler.HandleTextForChat(context.Background(), "Зупинити роботу", 77, "", Chat{ID: 10, Type: "group"}); err != nil {
		t.Fatal(err)
	}
	if flowRepo.state.Step != flows.StepStopWorkEnterPersonOrReason {
		t.Fatalf("stop flow not started: %#v", flowRepo.state)
	}
	got, err := handler.HandleTextForChat(context.Background(), "Іван Петренко готово", 77, "", Chat{ID: 10, Type: "group"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "зупинено") || flowRepo.ok || workRepo.records[0].Status != works.StatusStopped {
		t.Fatalf("stop flow not completed: %q ok=%v rec=%#v", got, flowRepo.ok, workRepo.records[0])
	}
}

func TestStartWorkButtonFlowManualDateStoresUTC(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	handler, personRepo, workRepo, flowRepo := newFlowHandler(now)
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Іван", LastName: "Петренко", NormalizedFullName: people.NormalizeFullName("Іван", "Петренко")}
	_, _ = personRepo.Create(context.Background(), p)
	inputs := []string{"Почати роботу", "Іван Петренко", "API", "Ввести дату вручну", "08.06 14:30"}
	var got string
	var err error
	for _, input := range inputs {
		got, err = handler.HandleTextForChat(context.Background(), input, 77, "", Chat{ID: 10, Type: "group"})
		if err != nil {
			t.Fatalf("%s: %v", input, err)
		}
	}
	if !strings.Contains(got, "розпочато") || flowRepo.ok || len(workRepo.records) != 1 {
		t.Fatalf("start flow not completed: %q ok=%v records=%#v", got, flowRepo.ok, workRepo.records)
	}
	if workRepo.records[0].StartedAt.Location() != time.UTC {
		t.Fatalf("started_at not UTC: %s", workRepo.records[0].StartedAt.Location())
	}
}
