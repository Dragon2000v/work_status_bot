package telegram

import (
	"context"
	"strings"
	"testing"
	"time"

	"work-status-bot/internal/i18n"
	"work-status-bot/internal/people"
	"work-status-bot/internal/reports"
	"work-status-bot/internal/users"
	"work-status-bot/internal/works"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newUserServiceForTelegramTests() *users.Service {
	repo := &memoryUserSettingsRepo{settings: map[int64]users.UserSetting{}}
	svc := users.NewService(repo)
	svc.SetNow(func() time.Time { return time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC) })
	return svc
}

type memoryUserSettingsRepo struct {
	settings map[int64]users.UserSetting
}

func (r *memoryUserSettingsRepo) GetByTelegramUserID(ctx context.Context, telegramUserID int64) (users.UserSetting, error) {
	setting, ok := r.settings[telegramUserID]
	if !ok {
		return users.UserSetting{}, users.ErrNotFound
	}
	return setting, nil
}

func (r *memoryUserSettingsRepo) UpsertLanguage(ctx context.Context, telegramUserID int64, language string, now time.Time) (users.UserSetting, error) {
	setting := r.settings[telegramUserID]
	if setting.CreatedAt.IsZero() {
		setting.CreatedAt = now
	}
	setting.TelegramUserID = telegramUserID
	setting.Language = language
	setting.UpdatedAt = now
	r.settings[telegramUserID] = setting
	return setting, nil
}

func TestLanguageCallbacksChangeOnlyPressingUser(t *testing.T) {
	userSvc := newUserServiceForTelegramTests()
	handler := NewHandlerWithUsersAndLogger(10, nil, nil, nil, userSvc, &fullTelegram{}, nil)
	got, _, outcome, err := handler.HandleCallback(context.Background(), CallbackLanguageEN, 1, 0)
	if err != nil || outcome != "success" || !strings.Contains(got, "Language saved") {
		t.Fatalf("en change got=%q outcome=%s err=%v", got, outcome, err)
	}
	lang1, _ := userSvc.LookupLanguage(context.Background(), 1)
	lang2, _ := userSvc.LookupLanguage(context.Background(), 2)
	if lang1 != i18n.English || lang2 != i18n.Ukrainian {
		t.Fatalf("langs %s %s", lang1, lang2)
	}
}

func TestPerUserLanguageResponsesDiffer(t *testing.T) {
	userSvc := newUserServiceForTelegramTests()
	_, _ = userSvc.SetLanguage(context.Background(), 1, i18n.Russian)
	_, _ = userSvc.SetLanguage(context.Background(), 2, i18n.English)
	handler := NewHandlerWithUsersAndLogger(10, nil, nil, nil, userSvc, &fullTelegram{}, nil)
	ru, _, _, _ := handler.HandleCallback(context.Background(), CallbackMenuHelp, 1, 0)
	en, _, _, _ := handler.HandleCallback(context.Background(), CallbackMenuHelp, 2, 0)
	if !strings.Contains(ru, "Команды") || !strings.Contains(en, "Commands") {
		t.Fatalf("ru=%q en=%q", ru, en)
	}
}

func TestTextCommandUsesSavedUserLanguage(t *testing.T) {
	userSvc := newUserServiceForTelegramTests()
	_, _ = userSvc.SetLanguage(context.Background(), 1, i18n.English)
	personRepo := newTelegramPeopleRepo()
	workRepo := &telegramWorkRepo{}
	handler := NewHandlerWithUsersAndLogger(10, fakePeopleService{personRepo}, fakeWorkService{workRepo, personRepo, time.Now().UTC()}, nil, userSvc, &fullTelegram{}, nil)
	got, err := handler.HandleTextForUser(context.Background(), "/help", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Commands") {
		t.Fatalf("want English help, got %q", got)
	}
}

func TestPublicAlertAndReportStayUkrainian(t *testing.T) {
	userSvc := newUserServiceForTelegramTests()
	_, _ = userSvc.SetLanguage(context.Background(), 1, i18n.English)
	p := people.Person{ID: primitive.NewObjectID(), FirstName: "Ivan", LastName: "Petrenko", NormalizedFullName: people.NormalizeFullName("Ivan", "Petrenko")}
	stopped := time.Date(2026, 6, 8, 11, 0, 0, 0, time.UTC)
	record := works.WorkRecord{ID: primitive.NewObjectID(), PersonID: p.ID, Title: "API", Status: works.StatusStopped, StartedAt: stopped.Add(-time.Hour), StoppedAt: &stopped, StopReason: "done"}
	if !strings.Contains(StopAlertMessage(p, record), "Увага:") {
		t.Fatalf("alert not Ukrainian: %q", StopAlertMessage(p, record))
	}
	report := ReportMessage(reports.GenerateResult{Duplicate: true, Report: reports.MonthlyReport{Content: "Місячний звіт 2026-06"}})
	if !strings.Contains(report, "Звіт уже існує") {
		t.Fatalf("report not Ukrainian: %q", report)
	}
}
