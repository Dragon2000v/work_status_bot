package flows

import (
	"testing"
	"time"
)

func TestNewStateSetsUTCAndExpiration(t *testing.T) {
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.FixedZone("x", 3*3600))
	state := NewState(1, 2, TypeStartWork, StepStartWorkEnterPerson, map[string]string{"a": "b"}, now)
	if state.UpdatedAt.Location() != time.UTC || state.CreatedAt.Location() != time.UTC {
		t.Fatalf("timestamps must be UTC: %#v", state)
	}
	if !state.ExpiresAt.Equal(state.UpdatedAt.Add(15 * time.Minute)) {
		t.Fatalf("expires_at = %s, want %s", state.ExpiresAt, state.UpdatedAt.Add(15*time.Minute))
	}
}

func TestStateWithStepRefreshesExpiration(t *testing.T) {
	state := NewState(1, 2, TypeStartWork, StepStartWorkEnterPerson, nil, time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC))
	advanced := state.WithStep(StepStartWorkEnterTitle, map[string]string{"first_name": "Ivan"}, state.UpdatedAt.Add(5*time.Minute))
	if advanced.Step != StepStartWorkEnterTitle {
		t.Fatalf("step = %s", advanced.Step)
	}
	if !advanced.ExpiresAt.Equal(advanced.UpdatedAt.Add(Expiration)) {
		t.Fatalf("expiration not refreshed")
	}
}
