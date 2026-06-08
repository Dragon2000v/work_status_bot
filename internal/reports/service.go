package reports

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"work-status-bot/internal/people"
	"work-status-bot/internal/works"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkLister interface {
	ListOverlappingMonth(ctx context.Context, month string) ([]works.WorkRecord, error)
}

type TelegramSender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type Service struct {
	repo      Repository
	people    people.Repository
	work      WorkLister
	telegram  TelegramSender
	groupChat int64
	now       func() time.Time
}

func NewService(repo Repository, peopleRepo people.Repository, work WorkLister, telegram TelegramSender, groupChat int64) *Service {
	return &Service{repo: repo, people: peopleRepo, work: work, telegram: telegram, groupChat: groupChat, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) SetNow(now func() time.Time) {
	s.now = now
}

func (s *Service) RecordStopped(ctx context.Context, personID, workRecordID primitive.ObjectID, reason string, occurredAt time.Time) error {
	return s.createIncident(ctx, IncidentStopped, personID, workRecordID, reason, occurredAt)
}

func (s *Service) RecordAlertFailed(ctx context.Context, personID, workRecordID primitive.ObjectID, reason string, occurredAt time.Time) error {
	return s.createIncident(ctx, IncidentAlertFailed, personID, workRecordID, reason, occurredAt)
}

func (s *Service) createIncident(ctx context.Context, typ string, personID, workRecordID primitive.ObjectID, reason string, occurredAt time.Time) error {
	now := s.now().UTC()
	return s.repo.CreateIncident(ctx, Incident{
		ID: primitive.NewObjectID(), PersonID: personID, WorkRecordID: workRecordID,
		Type: typ, Reason: strings.TrimSpace(reason), OccurredAt: occurredAt.UTC(), CreatedAt: now,
	})
}

type GenerateResult struct {
	Report    MonthlyReport
	Duplicate bool
}

func (s *Service) GenerateMonthly(ctx context.Context, month, source string) (GenerateResult, error) {
	if month == "" {
		month = works.CurrentMonthKyiv(s.now())
	}
	start, end, err := works.MonthBoundsKyiv(month)
	if err != nil {
		return GenerateResult{}, err
	}
	if existing, err := s.repo.FindMonthlyReport(ctx, month); err == nil {
		return GenerateResult{Report: existing, Duplicate: true}, nil
	} else if !errors.Is(err, ErrReportNotFound) {
		return GenerateResult{}, err
	}
	persons, err := s.people.List(ctx)
	if err != nil {
		return GenerateResult{}, err
	}
	records, err := s.work.ListOverlappingMonth(ctx, month)
	if err != nil {
		return GenerateResult{}, err
	}
	incidents, err := s.repo.ListIncidentsByMonth(ctx, start, end)
	if err != nil {
		return GenerateResult{}, err
	}
	content := RenderMonthlyReport(month, persons, records, incidents, s.now())
	now := s.now().UTC()
	report := MonthlyReport{ID: primitive.NewObjectID(), Month: month, TriggerSource: source, Content: content, GeneratedAt: now, CreatedAt: now}
	report, err = s.repo.CreateMonthlyReport(ctx, report)
	if errors.Is(err, ErrReportExists) {
		existing, findErr := s.repo.FindMonthlyReport(ctx, month)
		if findErr != nil {
			return GenerateResult{}, findErr
		}
		return GenerateResult{Report: existing, Duplicate: true}, nil
	}
	return GenerateResult{Report: report}, err
}

func (s *Service) SendReportToGroup(ctx context.Context, report MonthlyReport, duplicate bool) error {
	if s.telegram == nil || s.groupChat == 0 {
		return nil
	}
	prefix := ""
	if duplicate {
		prefix = "Report already exists.\n"
	}
	return s.telegram.SendMessage(ctx, s.groupChat, prefix+report.Content)
}

func RenderMonthlyReport(month string, persons []people.Person, records []works.WorkRecord, incidents []Incident, now time.Time) string {
	sort.Slice(persons, func(i, j int) bool { return persons[i].NormalizedFullName < persons[j].NormalizedFullName })
	byPerson := map[primitive.ObjectID]people.Person{}
	for _, p := range persons {
		byPerson[p.ID] = p
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Monthly report %s\n\nPeople:\n", month)
	for _, p := range persons {
		fmt.Fprintf(&b, "- %s %s\n", p.FirstName, p.LastName)
	}
	b.WriteString("\nWork records:\n")
	if len(records) == 0 {
		b.WriteString("- none\n")
	}
	for _, r := range records {
		p := byPerson[r.PersonID]
		fmt.Fprintf(&b, "- %s %s: %s, %s, started %s, elapsed %s\n", p.FirstName, p.LastName, r.Title, r.Status, works.FormatKyiv(r.StartedAt), works.FormatDuration(works.Elapsed(r, now)))
	}
	b.WriteString("\nIncidents:\n")
	if len(incidents) == 0 {
		b.WriteString("- none\n")
	}
	for _, i := range incidents {
		p := byPerson[i.PersonID]
		fmt.Fprintf(&b, "- %s %s: %s at %s", p.FirstName, p.LastName, i.Type, works.FormatKyiv(i.OccurredAt))
		if i.Reason != "" {
			fmt.Fprintf(&b, " (%s)", i.Reason)
		}
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
