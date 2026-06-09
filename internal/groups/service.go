package groups

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	OutcomeCreated         = "created"
	OutcomeUpdated         = "updated"
	OutcomeReEnabled       = "re_enabled"
	OutcomeDisabled        = "disabled"
	OutcomeAlreadyDisabled = "already_disabled"
)

type Service struct {
	repo         Repository
	fallbackChat int64
	now          func() time.Time
}

func NewService(repo Repository, fallbackChat int64) *Service {
	return &Service{repo: repo, fallbackChat: fallbackChat, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) SetNow(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func (s *Service) Setup(ctx context.Context, req GroupSetupRequest) (SetupResult, error) {
	if s == nil || s.repo == nil {
		return SetupResult{}, errors.New("group repository not configured")
	}
	if req.TelegramChatID == 0 {
		return SetupResult{}, errors.New("telegram chat id is required")
	}
	if req.SetupUserID == 0 {
		return SetupResult{}, errors.New("setup user id is required")
	}
	now := s.now().UTC()
	if !req.OccurredAt.IsZero() {
		now = req.OccurredAt.UTC()
	}
	req.Title = NormalizeTitle(req.Title)
	group, existed, wasDisabled, err := s.repo.UpsertSetup(ctx, req, now)
	if err != nil {
		return SetupResult{}, err
	}
	outcome := OutcomeCreated
	if wasDisabled {
		outcome = OutcomeReEnabled
	} else if existed {
		outcome = OutcomeUpdated
	}
	return SetupResult{Group: group, Outcome: outcome}, nil
}

func (s *Service) Disable(ctx context.Context, chatID int64) (ConfiguredGroup, string, error) {
	if s == nil || s.repo == nil {
		return ConfiguredGroup{}, "", errors.New("group repository not configured")
	}
	group, alreadyDisabled, err := s.repo.Disable(ctx, chatID, s.now().UTC())
	if err != nil {
		return ConfiguredGroup{}, "", err
	}
	if alreadyDisabled {
		return group, OutcomeAlreadyDisabled, nil
	}
	return group, OutcomeDisabled, nil
}

func (s *Service) ListAll(ctx context.Context) ([]ConfiguredGroup, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	return s.repo.ListAll(ctx)
}

func (s *Service) ListEnabled(ctx context.Context) ([]ConfiguredGroup, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	return s.repo.ListEnabled(ctx)
}

func (s *Service) CronTargets(ctx context.Context) ([]ReportDeliveryTarget, error) {
	var targets []ReportDeliveryTarget
	seen := map[int64]bool{}
	groups, err := s.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if group.TelegramChatID == 0 || seen[group.TelegramChatID] {
			continue
		}
		seen[group.TelegramChatID] = true
		targets = append(targets, ReportDeliveryTarget{TelegramChatID: group.TelegramChatID, Title: group.Title, Source: ReportTargetConfiguredGroup})
	}
	if s != nil && s.fallbackChat != 0 && !seen[s.fallbackChat] {
		targets = append(targets, ReportDeliveryTarget{TelegramChatID: s.fallbackChat, Title: "Fallback group", Source: ReportTargetFallbackGroup})
	}
	return targets, nil
}

func (s *Service) Authorize(ctx context.Context, chatID int64, chatType, command string) GroupAuthorizationDecision {
	decision := GroupAuthorizationDecision{ChatID: chatID, Command: command}
	if chatType == "private" {
		if command == "/setup" {
			decision.Reason = ReasonPrivateChatSetup
		} else {
			decision.Reason = ReasonPrivateChat
		}
		return decision
	}
	if s != nil && s.fallbackChat != 0 && chatID == s.fallbackChat {
		decision.Allowed = true
		decision.Reason = ReasonFallback
		return decision
	}
	if s != nil && s.repo != nil && chatID != 0 {
		group, err := s.repo.FindByTelegramChatID(ctx, chatID)
		if err == nil {
			if group.Enabled {
				decision.Allowed = true
				decision.Reason = ReasonConfigured
				return decision
			}
			if command == "/setup" {
				decision.Allowed = true
				decision.Reason = ReasonSetupAllowed
				return decision
			}
			if command == "/help" {
				decision.Allowed = true
				decision.Reason = ReasonHelpAllowed
				return decision
			}
			decision.Reason = ReasonDisabledGroup
			return decision
		}
	}
	switch command {
	case "/setup":
		decision.Allowed = true
		decision.Reason = ReasonSetupAllowed
	case "/help":
		decision.Allowed = true
		decision.Reason = ReasonHelpAllowed
	default:
		decision.Reason = ReasonUnknownGroup
	}
	return decision
}

func NormalizeTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return UnknownGroupTitle
	}
	return title
}

func FormatGroupList(groups []ConfiguredGroup, fallbackChat int64) string {
	var b strings.Builder
	b.WriteString("Налаштовані групи:\n")
	if len(groups) == 0 {
		b.WriteString("- немає збережених груп\n")
		if fallbackChat != 0 {
			fmt.Fprintf(&b, "- Fallback group (%d): enabled\n", fallbackChat)
		}
		return strings.TrimSpace(b.String())
	}
	for _, group := range groups {
		status := "disabled"
		if group.Enabled {
			status = "enabled"
		}
		fmt.Fprintf(&b, "- %s (%d): %s\n", group.Title, group.TelegramChatID, status)
	}
	return strings.TrimSpace(b.String())
}
