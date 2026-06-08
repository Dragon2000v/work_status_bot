package users

import (
	"context"
	"errors"
	"time"

	"work-status-bot/internal/i18n"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) SetNow(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func (s *Service) LookupLanguage(ctx context.Context, telegramUserID int64) (i18n.Language, error) {
	if s == nil || s.repo == nil || telegramUserID == 0 {
		return i18n.Ukrainian, nil
	}
	setting, err := s.repo.GetByTelegramUserID(ctx, telegramUserID)
	if errors.Is(err, ErrNotFound) {
		return i18n.Ukrainian, nil
	}
	if err != nil {
		return i18n.Ukrainian, err
	}
	return i18n.Normalize(i18n.Language(setting.Language)), nil
}

func (s *Service) SetLanguage(ctx context.Context, telegramUserID int64, lang i18n.Language) (UserSetting, error) {
	if telegramUserID == 0 {
		return UserSetting{}, errors.New("telegram user id is required")
	}
	if !i18n.Supported(lang) {
		return UserSetting{}, errors.New("unsupported language")
	}
	return s.repo.UpsertLanguage(ctx, telegramUserID, string(lang), s.now().UTC())
}
