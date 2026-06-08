package reports

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrReportExists = errors.New("monthly report already exists")
var ErrReportNotFound = errors.New("monthly report not found")

type Repository interface {
	CreateIncident(ctx context.Context, incident Incident) error
	ListIncidentsByMonth(ctx context.Context, startUTC, endUTC time.Time) ([]Incident, error)
	CreateMonthlyReport(ctx context.Context, report MonthlyReport) (MonthlyReport, error)
	FindMonthlyReport(ctx context.Context, month string) (MonthlyReport, error)
}

type ReportsRepository struct {
	incidents *mongo.Collection
	reports   *mongo.Collection
}

func NewRepository(db *mongo.Database) *ReportsRepository {
	return &ReportsRepository{incidents: db.Collection("incidents"), reports: db.Collection("monthly_reports")}
}

func (r *ReportsRepository) CreateIncident(ctx context.Context, incident Incident) error {
	if incident.ID.IsZero() {
		incident.ID = primitive.NewObjectID()
	}
	_, err := r.incidents.InsertOne(ctx, incident)
	return err
}

func (r *ReportsRepository) ListIncidentsByMonth(ctx context.Context, startUTC, endUTC time.Time) ([]Incident, error) {
	cur, err := r.incidents.Find(ctx, bson.M{"occurred_at": bson.M{"$gte": startUTC.UTC(), "$lt": endUTC.UTC()}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []Incident
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ReportsRepository) CreateMonthlyReport(ctx context.Context, report MonthlyReport) (MonthlyReport, error) {
	_, err := r.reports.InsertOne(ctx, report)
	if mongo.IsDuplicateKeyError(err) {
		return MonthlyReport{}, ErrReportExists
	}
	return report, err
}

func (r *ReportsRepository) FindMonthlyReport(ctx context.Context, month string) (MonthlyReport, error) {
	var report MonthlyReport
	err := r.reports.FindOne(ctx, bson.M{"month": month}).Decode(&report)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return MonthlyReport{}, ErrReportNotFound
	}
	return report, err
}
