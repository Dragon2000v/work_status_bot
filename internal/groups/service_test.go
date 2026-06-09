package groups

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type memGroupRepo struct {
	groups map[int64]ConfiguredGroup
}

func newMemGroupRepo() *memGroupRepo {
	return &memGroupRepo{groups: map[int64]ConfiguredGroup{}}
}

func (r *memGroupRepo) UpsertSetup(ctx context.Context, req GroupSetupRequest, now time.Time) (ConfiguredGroup, bool, bool, error) {
	group, existed := r.groups[req.TelegramChatID]
	wasDisabled := existed && !group.Enabled
	if !existed {
		group = ConfiguredGroup{ID: primitive.NewObjectID(), TelegramChatID: req.TelegramChatID, CreatedAt: now.UTC()}
	}
	group.Title = NormalizeTitle(req.Title)
	group.SetupUserID = req.SetupUserID
	group.SetupUsername = req.SetupUsername
	group.Enabled = true
	group.UpdatedAt = now.UTC()
	r.groups[req.TelegramChatID] = group
	return group, existed, wasDisabled, nil
}

func (r *memGroupRepo) FindByTelegramChatID(ctx context.Context, chatID int64) (ConfiguredGroup, error) {
	group, ok := r.groups[chatID]
	if !ok {
		return ConfiguredGroup{}, ErrNotFound
	}
	return group, nil
}

func (r *memGroupRepo) ListEnabled(ctx context.Context) ([]ConfiguredGroup, error) {
	var out []ConfiguredGroup
	for _, group := range r.groups {
		if group.Enabled {
			out = append(out, group)
		}
	}
	return out, nil
}

func (r *memGroupRepo) ListAll(ctx context.Context) ([]ConfiguredGroup, error) {
	var out []ConfiguredGroup
	for _, group := range r.groups {
		out = append(out, group)
	}
	return out, nil
}

func (r *memGroupRepo) Disable(ctx context.Context, chatID int64, now time.Time) (ConfiguredGroup, bool, error) {
	group, ok := r.groups[chatID]
	if !ok {
		return ConfiguredGroup{}, false, ErrNotFound
	}
	alreadyDisabled := !group.Enabled
	group.Enabled = false
	group.UpdatedAt = now.UTC()
	r.groups[chatID] = group
	return group, alreadyDisabled, nil
}

func TestServiceSetupCreateUpdateReenableAndNormalizeTitle(t *testing.T) {
	repo := newMemGroupRepo()
	svc := NewService(repo, 99)
	first := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
	second := first.Add(time.Hour)
	svc.SetNow(func() time.Time { return first })
	result, err := svc.Setup(context.Background(), GroupSetupRequest{TelegramChatID: -1001, Title: "  ", SetupUserID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != OutcomeCreated || result.Group.Title != UnknownGroupTitle || !result.Group.Enabled {
		t.Fatalf("bad create: %#v", result)
	}
	createdAt := result.Group.CreatedAt

	svc.SetNow(func() time.Time { return second })
	result, err = svc.Setup(context.Background(), GroupSetupRequest{TelegramChatID: -1001, Title: " Team ", SetupUserID: 8, SetupUsername: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != OutcomeUpdated || result.Group.Title != "Team" || result.Group.SetupUserID != 8 || result.Group.SetupUsername != "admin" {
		t.Fatalf("bad update: %#v", result)
	}
	if !result.Group.CreatedAt.Equal(createdAt) || !result.Group.UpdatedAt.Equal(second) {
		t.Fatalf("timestamps not preserved/updated: %#v", result.Group)
	}
	if _, outcome, err := svc.Disable(context.Background(), -1001); err != nil || outcome != OutcomeDisabled {
		t.Fatalf("disable outcome=%s err=%v", outcome, err)
	}
	result, err = svc.Setup(context.Background(), GroupSetupRequest{TelegramChatID: -1001, Title: "Team", SetupUserID: 9})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome != OutcomeReEnabled || !result.Group.Enabled {
		t.Fatalf("want re-enabled: %#v", result)
	}
}

func TestServiceAuthorizeConfiguredFallbackUnknownDisabledAndPrivate(t *testing.T) {
	repo := newMemGroupRepo()
	svc := NewService(repo, -10099)
	_, err := svc.Setup(context.Background(), GroupSetupRequest{TelegramChatID: -1001, Title: "Team", SetupUserID: 7})
	if err != nil {
		t.Fatal(err)
	}
	repo.groups[-1002] = ConfiguredGroup{TelegramChatID: -1002, Title: "Disabled", Enabled: false}
	cases := []struct {
		name    string
		chatID  int64
		typ     string
		command string
		allowed bool
		reason  string
	}{
		{"configured", -1001, "group", "/status", true, ReasonConfigured},
		{"fallback", -10099, "group", "/status", true, ReasonFallback},
		{"unknown setup", -1003, "group", "/setup", true, ReasonSetupAllowed},
		{"unknown help", -1003, "group", "/help", true, ReasonHelpAllowed},
		{"unknown status", -1003, "group", "/status", false, ReasonUnknownGroup},
		{"disabled status", -1002, "group", "/status", false, ReasonDisabledGroup},
		{"disabled setup", -1002, "group", "/setup", true, ReasonSetupAllowed},
		{"disabled help", -1002, "group", "/help", true, ReasonHelpAllowed},
		{"private setup", 1, "private", "/setup", false, ReasonPrivateChatSetup},
		{"private groups", 1, "private", "/groups", false, ReasonPrivateChat},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := svc.Authorize(context.Background(), tc.chatID, tc.typ, tc.command)
			if got.Allowed != tc.allowed || got.Reason != tc.reason {
				t.Fatalf("got %#v", got)
			}
		})
	}
}

func TestServiceDisableIdempotentAndListTargets(t *testing.T) {
	repo := newMemGroupRepo()
	svc := NewService(repo, -1003)
	_, _ = svc.Setup(context.Background(), GroupSetupRequest{TelegramChatID: -1001, Title: "One", SetupUserID: 1})
	_, _ = svc.Setup(context.Background(), GroupSetupRequest{TelegramChatID: -1002, Title: "Two", SetupUserID: 1})
	if _, outcome, err := svc.Disable(context.Background(), -1002); err != nil || outcome != OutcomeDisabled {
		t.Fatalf("disable outcome=%s err=%v", outcome, err)
	}
	if _, outcome, err := svc.Disable(context.Background(), -1002); err != nil || outcome != OutcomeAlreadyDisabled {
		t.Fatalf("second disable outcome=%s err=%v", outcome, err)
	}
	all, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("want all groups including disabled, got %d", len(all))
	}
	targets, err := svc.CronTargets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 {
		t.Fatalf("want enabled group plus fallback, got %#v", targets)
	}
	for _, target := range targets {
		if target.TelegramChatID == -1002 {
			t.Fatalf("disabled group included in cron targets: %#v", targets)
		}
	}
}

func TestServiceDisableMissingGroup(t *testing.T) {
	_, _, err := NewService(newMemGroupRepo(), 0).Disable(context.Background(), -404)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
