package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"work-status-bot/internal/i18n"
)

type memoryRepo struct {
	setting UserSetting
	err     error
	writes  int
}

func (r *memoryRepo) GetByTelegramUserID(ctx context.Context, telegramUserID int64) (UserSetting, error) {
	if r.err != nil {
		return UserSetting{}, r.err
	}
	if r.setting.TelegramUserID == 0 {
		return UserSetting{}, ErrNotFound
	}
	return r.setting, nil
}

func (r *memoryRepo) UpsertLanguage(ctx context.Context, telegramUserID int64, language string, now time.Time) (UserSetting, error) {
	r.writes++
	if r.setting.CreatedAt.IsZero() {
		r.setting.CreatedAt = now
	}
	r.setting.TelegramUserID = telegramUserID
	r.setting.Language = language
	r.setting.UpdatedAt = now
	return r.setting, nil
}

func TestLookupLanguageFallbackAndSaved(t *testing.T) {
	svc := NewService(&memoryRepo{})
	got, err := svc.LookupLanguage(context.Background(), 1)
	if err != nil || got != i18n.Ukrainian {
		t.Fatalf("missing fallback got %s err %v", got, err)
	}
	svc = NewService(&memoryRepo{setting: UserSetting{TelegramUserID: 1, Language: "en"}})
	got, err = svc.LookupLanguage(context.Background(), 1)
	if err != nil || got != i18n.English {
		t.Fatalf("saved got %s err %v", got, err)
	}
	svc = NewService(&memoryRepo{setting: UserSetting{TelegramUserID: 1, Language: "bad"}})
	got, err = svc.LookupLanguage(context.Background(), 1)
	if err != nil || got != i18n.Ukrainian {
		t.Fatalf("bad fallback got %s err %v", got, err)
	}
}

func TestSetLanguageValidationAndWrite(t *testing.T) {
	repo := &memoryRepo{}
	svc := NewService(repo)
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	svc.SetNow(func() time.Time { return now })
	if _, err := svc.SetLanguage(context.Background(), 1, i18n.Language("pl")); err == nil {
		t.Fatalf("want unsupported language error")
	}
	setting, err := svc.SetLanguage(context.Background(), 1, i18n.Russian)
	if err != nil {
		t.Fatal(err)
	}
	if setting.Language != "ru" || repo.writes != 1 || !setting.UpdatedAt.Equal(now) {
		t.Fatalf("bad setting %#v writes %d", setting, repo.writes)
	}
}

func TestLookupLanguageReturnsRepositoryError(t *testing.T) {
	svc := NewService(&memoryRepo{err: errors.New("db failed")})
	if _, err := svc.LookupLanguage(context.Background(), 1); err == nil {
		t.Fatalf("want repo error")
	}
}
