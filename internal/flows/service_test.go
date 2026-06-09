package flows

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memRepo struct {
	state State
	ok    bool
}

func (r *memRepo) Upsert(ctx context.Context, state State) (State, error) {
	r.state = state
	r.ok = true
	return state, nil
}

func (r *memRepo) Get(ctx context.Context, telegramUserID, chatID int64) (State, error) {
	if !r.ok {
		return State{}, ErrNotFound
	}
	return r.state, nil
}

func (r *memRepo) Delete(ctx context.Context, telegramUserID, chatID int64) error {
	r.ok = false
	return nil
}

func TestServiceStartAdvanceCompleteAndExpire(t *testing.T) {
	repo := &memRepo{}
	svc := NewService(repo)
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	svc.SetNow(func() time.Time { return now })
	state, err := svc.Start(context.Background(), 1, 2, TypeSettings, StepSettingsLanguageSelect, nil)
	if err != nil {
		t.Fatal(err)
	}
	if state.Step != StepSettingsLanguageSelect || !repo.ok {
		t.Fatalf("state not started: %#v", state)
	}
	now = now.Add(time.Minute)
	state, err = svc.Advance(context.Background(), state, StepSettingsLanguageSelect, map[string]string{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if state.Payload["x"] != "y" || !state.ExpiresAt.Equal(now.Add(Expiration)) {
		t.Fatalf("state not advanced: %#v", state)
	}
	now = now.Add(Expiration)
	if _, err := svc.Get(context.Background(), 1, 2); !errors.Is(err, ErrExpired) {
		t.Fatalf("want expired, got %v", err)
	}
	if repo.ok {
		t.Fatalf("expired state not cleared")
	}
}
