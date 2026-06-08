package works

import (
	"fmt"
	"time"
)

var KyivLocation = mustKyiv()

func mustKyiv() *time.Location {
	loc, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		return time.FixedZone("Europe/Kyiv", 2*60*60)
	}
	return loc
}

func FormatKyiv(t time.Time) string {
	return t.UTC().In(KyivLocation).Format("2006-01-02 15:04")
}

func Elapsed(record WorkRecord, now time.Time) time.Duration {
	end := now.UTC()
	if record.StoppedAt != nil {
		end = record.StoppedAt.UTC()
	}
	if end.Before(record.StartedAt.UTC()) {
		return 0
	}
	return end.Sub(record.StartedAt.UTC())
}

func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	totalMinutes := int(d.Round(time.Minute).Minutes())
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	if hours == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func MonthBoundsKyiv(month string) (time.Time, time.Time, error) {
	startLocal, err := time.ParseInLocation("2006-01", month, KyivLocation)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return startLocal.UTC(), startLocal.AddDate(0, 1, 0).UTC(), nil
}

func CurrentMonthKyiv(now time.Time) string {
	return now.UTC().In(KyivLocation).Format("2006-01")
}
