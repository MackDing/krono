package schedule_test

import (
	"testing"
	"time"

	"github.com/MackDing/krono/internal/schedule"
)

func TestInterval(t *testing.T) {
	s, err := schedule.Parse("@every 30s")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	base := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	if got, want := s.Next(base), base.Add(30*time.Second); !got.Equal(want) {
		t.Errorf("Next = %v, want %v", got, want)
	}
}

func TestCronNext(t *testing.T) {
	cases := []struct {
		spec, from, want string
	}{
		{"0 8 * * *", "2026-05-18T06:00:00Z", "2026-05-18T08:00:00Z"},   // later today
		{"0 8 * * *", "2026-05-18T09:00:00Z", "2026-05-19T08:00:00Z"},   // already past -> tomorrow
		{"*/15 * * * *", "2026-05-18T12:07:00Z", "2026-05-18T12:15:00Z"}, // step
		{"0 0 1 * *", "2026-05-18T12:00:00Z", "2026-06-01T00:00:00Z"},   // 1st of next month
		{"0 8 * * 1-5", "2026-05-16T10:00:00Z", "2026-05-18T08:00:00Z"}, // Sat -> Mon (weekdays)
		{"@daily", "2026-05-18T12:00:00Z", "2026-05-19T00:00:00Z"},      // shortcut
		{"30 9 * * *", "2026-05-18T09:30:00Z", "2026-05-19T09:30:00Z"},  // exactly on -> strictly after
	}
	for _, c := range cases {
		s, err := schedule.Parse(c.spec)
		if err != nil {
			t.Errorf("%s: parse: %v", c.spec, err)
			continue
		}
		from, _ := time.Parse(time.RFC3339, c.from)
		want, _ := time.Parse(time.RFC3339, c.want)
		if got := s.Next(from); !got.Equal(want) {
			t.Errorf("%s: Next(%s) = %s, want %s",
				c.spec, c.from, got.Format(time.RFC3339), c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	bad := []string{
		"", "0 8 * *", "0 8 * * * *", "60 8 * * *", "0 24 * * *",
		"0 8 0 * *", "0 8 * 13 *", "0 8 * * 8", "xyz", "@every", "@every 0s", "@every -5s",
	}
	for _, spec := range bad {
		if _, err := schedule.Parse(spec); err == nil {
			t.Errorf("Parse(%q) = nil error, want an error", spec)
		}
	}
}
