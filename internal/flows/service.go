package flows

import (
	"context"
	"errors"
	"time"
)

var ErrExpired = errors.New("flow state expired")

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

func (s *Service) Start(ctx context.Context, userID, chatID int64, flowType, step string, payload map[string]string) (State, error) {
	return s.repo.Upsert(ctx, NewState(userID, chatID, flowType, step, payload, s.now()))
}

func (s *Service) Get(ctx context.Context, userID, chatID int64) (State, error) {
	state, err := s.repo.Get(ctx, userID, chatID)
	if err != nil {
		return State{}, err
	}
	if state.Expired(s.now()) {
		_ = s.repo.Delete(ctx, userID, chatID)
		return State{}, ErrExpired
	}
	return state, nil
}

func (s *Service) Advance(ctx context.Context, state State, step string, payload map[string]string) (State, error) {
	return s.repo.Upsert(ctx, state.WithStep(step, payload, s.now()))
}

func (s *Service) Complete(ctx context.Context, userID, chatID int64) error {
	return s.repo.Delete(ctx, userID, chatID)
}

func (s *Service) Cancel(ctx context.Context, userID, chatID int64) error {
	return s.repo.Delete(ctx, userID, chatID)
}
