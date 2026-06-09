package telegram

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"work-status-bot/internal/groups"
	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/works"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type telegramGroupRepo struct {
	groups map[int64]groups.ConfiguredGroup
}

func newTelegramGroupRepo() *telegramGroupRepo {
	return &telegramGroupRepo{groups: map[int64]groups.ConfiguredGroup{}}
}

func (r *telegramGroupRepo) UpsertSetup(ctx context.Context, req groups.GroupSetupRequest, now time.Time) (groups.ConfiguredGroup, bool, bool, error) {
	group, existed := r.groups[req.TelegramChatID]
	wasDisabled := existed && !group.Enabled
	if !existed {
		group = groups.ConfiguredGroup{ID: primitive.NewObjectID(), TelegramChatID: req.TelegramChatID, CreatedAt: now.UTC()}
	}
	group.Title = groups.NormalizeTitle(req.Title)
	group.SetupUserID = req.SetupUserID
	group.SetupUsername = req.SetupUsername
	group.Enabled = true
	group.UpdatedAt = now.UTC()
	r.groups[req.TelegramChatID] = group
	return group, existed, wasDisabled, nil
}

func (r *telegramGroupRepo) FindByTelegramChatID(ctx context.Context, chatID int64) (groups.ConfiguredGroup, error) {
	group, ok := r.groups[chatID]
	if !ok {
		return groups.ConfiguredGroup{}, groups.ErrNotFound
	}
	return group, nil
}

func (r *telegramGroupRepo) ListEnabled(ctx context.Context) ([]groups.ConfiguredGroup, error) {
	var out []groups.ConfiguredGroup
	for _, group := range r.groups {
		if group.Enabled {
			out = append(out, group)
		}
	}
	return out, nil
}

func (r *telegramGroupRepo) ListAll(ctx context.Context) ([]groups.ConfiguredGroup, error) {
	var out []groups.ConfiguredGroup
	for _, group := range r.groups {
		out = append(out, group)
	}
	return out, nil
}

func (r *telegramGroupRepo) Disable(ctx context.Context, chatID int64, now time.Time) (groups.ConfiguredGroup, bool, error) {
	group, ok := r.groups[chatID]
	if !ok {
		return groups.ConfiguredGroup{}, false, groups.ErrNotFound
	}
	alreadyDisabled := !group.Enabled
	group.Enabled = false
	group.UpdatedAt = now.UTC()
	r.groups[chatID] = group
	return group, alreadyDisabled, nil
}

func newGroupHandler(repo *telegramGroupRepo, fallback int64, tg interface {
	SendMessage(context.Context, int64, string) error
}) *Handler {
	personRepo := newTelegramPeopleRepo()
	workRepo := &telegramWorkRepo{}
	return NewHandlerWithGroupsUsersAndLogger(fallback, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, time.Now().UTC()}, nil, groups.NewService(repo, fallback), nil, tg, nil)
}

func TestSetupCommandGroupPrivateIdempotentAndMissingTitle(t *testing.T) {
	repo := newTelegramGroupRepo()
	handler := newGroupHandler(repo, 10, nil)
	got, err := handler.HandleTextForChat(context.Background(), "/setup", 77, "admin", Chat{ID: -1001, Type: "group"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "завершено") || repo.groups[-1001].Title != groups.UnknownGroupTitle {
		t.Fatalf("bad setup response=%q group=%#v", got, repo.groups[-1001])
	}
	got, err = handler.HandleTextForChat(context.Background(), "/setup", 78, "", Chat{ID: -1001, Type: "group", Title: "Team"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "вже було") || repo.groups[-1001].Title != "Team" {
		t.Fatalf("bad idempotent setup response=%q group=%#v", got, repo.groups[-1001])
	}
	got, err = handler.HandleTextForChat(context.Background(), "/setup", 77, "", Chat{ID: 1, Type: "private"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "лише в групових") {
		t.Fatalf("want private rejection, got %q", got)
	}
}

func TestGroupsCommandAvailabilityAndOutput(t *testing.T) {
	repo := newTelegramGroupRepo()
	handler := newGroupHandler(repo, 10, nil)
	_, _ = handler.HandleTextForChat(context.Background(), "/setup", 1, "", Chat{ID: -1001, Type: "group", Title: "Enabled"})
	_, _ = handler.HandleTextForChat(context.Background(), "/setup", 1, "", Chat{ID: -1002, Type: "group", Title: "Disabled"})
	_, _ = handler.HandleTextForChat(context.Background(), "/disable_group", 1, "", Chat{ID: -1002, Type: "group", Title: "Disabled"})

	got, err := handler.HandleTextForChat(context.Background(), "/groups", 1, "", Chat{ID: -1001, Type: "group"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Enabled", "Disabled", "enabled", "disabled"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %s in %q", want, got)
		}
	}
	got, _ = handler.HandleTextForChat(context.Background(), "/groups", 1, "", Chat{ID: 10, Type: "group"})
	if !strings.Contains(got, "Enabled") || !strings.Contains(got, "Disabled") {
		t.Fatalf("fallback should list stored groups: %q", got)
	}
	for _, chat := range []Chat{{ID: -1002, Type: "group"}, {ID: -9999, Type: "group"}, {ID: 5, Type: "private"}} {
		got, err = handler.HandleTextForChat(context.Background(), "/groups", 1, "", chat)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got, "не налаштовано") {
			t.Fatalf("want rejection for %#v, got %q", chat, got)
		}
	}
}

func TestUnknownGroupSetupAllowedAndOperationalRejected(t *testing.T) {
	repo := newTelegramGroupRepo()
	handler := newGroupHandler(repo, 10, nil)
	got, err := handler.HandleTextForChat(context.Background(), "/status", 1, "", Chat{ID: -900, Type: "group"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "не налаштовано") {
		t.Fatalf("want setup-required rejection, got %q", got)
	}
	got, err = handler.HandleTextForChat(context.Background(), "/setup", 1, "", Chat{ID: -900, Type: "group", Title: "New"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "завершено") || !repo.groups[-900].Enabled {
		t.Fatalf("setup not allowed from unknown group: %q %#v", got, repo.groups[-900])
	}
}

func TestCallbackAcknowledgedAfterGroupAuthorizationRejection(t *testing.T) {
	repo := newTelegramGroupRepo()
	tg := &fullTelegram{}
	handler := newGroupHandler(repo, 10, tg)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"callback_query":{"id":"cb1","from":{"id":77},"message":{"message_id":9,"chat":{"id":-999}},"data":"menu:status"}}`)))
	if res.Code != http.StatusOK || len(tg.answers) != 1 {
		t.Fatalf("code=%d answers=%v", res.Code, tg.answers)
	}
	if len(tg.messages) != 1 || !strings.Contains(tg.messages[0], "не налаштовано") {
		t.Fatalf("want setup-required callback response, got %#v", tg.messages)
	}
}

type chatRecordingTelegram struct {
	chats []int64
	texts []string
}

func (t *chatRecordingTelegram) SendMessage(ctx context.Context, chatID int64, text string) error {
	t.chats = append(t.chats, chatID)
	t.texts = append(t.texts, text)
	return nil
}

func TestStopAlertGoesToCurrentGroup(t *testing.T) {
	repo := newTelegramGroupRepo()
	_, _, _, _ = repo.UpsertSetup(context.Background(), groups.GroupSetupRequest{TelegramChatID: -1001, Title: "Team", SetupUserID: 1}, time.Now())
	personRepo := newTelegramPeopleRepo()
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")}
	_, _ = personRepo.Create(context.Background(), p)
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	workRepo := &telegramWorkRepo{records: []works.WorkRecord{{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "API", Status: works.StatusActive, StartedAt: now}}}
	reportRepo := newTelegramReportsRepo()
	reportSvc := reports.NewService(reportRepo, personRepo, telegramReportWorkLister{}, nil, 0)
	tg := &chatRecordingTelegram{}
	handler := NewHandlerWithGroupsUsersAndLogger(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, now}, reportSvc, groups.NewService(repo, 10), nil, tg, nil)
	if _, err := handler.HandleTextForChat(context.Background(), "/stop_work Ivan Petrenko done", 1, "", Chat{ID: -1001, Type: "group"}); err != nil {
		t.Fatal(err)
	}
	if len(tg.chats) != 1 || tg.chats[0] != -1001 {
		t.Fatalf("alert should target current group, got chats=%v texts=%v", tg.chats, tg.texts)
	}
}
