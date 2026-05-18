// Package schedule parses cron expressions and interval specs into Schedule
// values that compute the next time a job should fire.
package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schedule computes the next time a job should run.
type Schedule interface {
	// Next returns the first activation strictly after t, in t's location.
	// It returns the zero Time if the schedule will never fire again.
	Next(t time.Time) time.Time
	// String returns the spec the schedule was parsed from.
	String() string
}

var shortcuts = map[string]string{
	"@hourly":   "0 * * * *",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@weekly":   "0 0 * * 0",
	"@monthly":  "0 0 1 * *",
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
}

// Parse converts a spec string into a Schedule. Accepted forms:
//
//	@every <duration>                  e.g. "@every 30s", "@every 1h30m"
//	@hourly @daily @weekly @monthly @yearly
//	<min> <hour> <dom> <mon> <dow>     standard 5-field cron, e.g. "0 8 * * 1-5"
func Parse(spec string) (Schedule, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("schedule: empty spec")
	}
	if rest, ok := strings.CutPrefix(spec, "@every "); ok {
		d, err := time.ParseDuration(strings.TrimSpace(rest))
		if err != nil {
			return nil, fmt.Errorf("schedule: invalid @every duration %q: %w", rest, err)
		}
		if d <= 0 {
			return nil, fmt.Errorf("schedule: @every duration must be positive, got %s", d)
		}
		return &Interval{every: d, spec: spec}, nil
	}
	cronSpec := spec
	if sc, ok := shortcuts[spec]; ok {
		cronSpec = sc
	}
	return parseCron(spec, cronSpec)
}

// Interval fires at a fixed period.
type Interval struct {
	every time.Duration
	spec  string
}

// Next returns t advanced by the interval.
func (i *Interval) Next(t time.Time) time.Time { return t.Add(i.every) }

// String returns the original "@every ..." spec.
func (i *Interval) String() string { return i.spec }

// Every returns the interval duration.
func (i *Interval) Every() time.Duration { return i.every }

// CronSchedule is a parsed standard 5-field cron expression.
type CronSchedule struct {
	minute, hour, dom, month, dow uint64
	domStar, dowStar              bool
	spec                          string
}

// String returns the original cron spec.
func (c *CronSchedule) String() string { return c.spec }

func parseCron(orig, spec string) (*CronSchedule, error) {
	fields := strings.Fields(spec)
	if len(fields) != 5 {
		return nil, fmt.Errorf("schedule: cron expression %q must have 5 fields, got %d", orig, len(fields))
	}
	c := &CronSchedule{spec: orig}
	var err error
	if c.minute, _, err = parseField(fields[0], 0, 59); err != nil {
		return nil, fmt.Errorf("schedule: minute field: %w", err)
	}
	if c.hour, _, err = parseField(fields[1], 0, 23); err != nil {
		return nil, fmt.Errorf("schedule: hour field: %w", err)
	}
	if c.dom, c.domStar, err = parseField(fields[2], 1, 31); err != nil {
		return nil, fmt.Errorf("schedule: day-of-month field: %w", err)
	}
	if c.month, _, err = parseField(fields[3], 1, 12); err != nil {
		return nil, fmt.Errorf("schedule: month field: %w", err)
	}
	if c.dow, c.dowStar, err = parseField(fields[4], 0, 7); err != nil {
		return nil, fmt.Errorf("schedule: day-of-week field: %w", err)
	}
	// Cron allows both 0 and 7 to mean Sunday; fold bit 7 onto bit 0.
	if c.dow&(1<<7) != 0 {
		c.dow = (c.dow &^ (1 << 7)) | 1
	}
	return c, nil
}

// parseField parses one cron field into a bitmask over [min,max]. The bool
// reports whether the field was the unrestricted "*".
func parseField(field string, min, max int) (uint64, bool, error) {
	if field == "*" {
		return fullMask(min, max), true, nil
	}
	var mask uint64
	for _, part := range strings.Split(field, ",") {
		m, err := parseTerm(part, min, max)
		if err != nil {
			return 0, false, err
		}
		mask |= m
	}
	return mask, false, nil
}

// parseTerm parses one comma-separated term: "n", "a-b", "*/s", "a-b/s", "n/s".
func parseTerm(part string, min, max int) (uint64, error) {
	rng, stepStr, hasStep := strings.Cut(part, "/")
	step := 1
	if hasStep {
		s, err := strconv.Atoi(stepStr)
		if err != nil || s < 1 {
			return 0, fmt.Errorf("invalid step in %q", part)
		}
		step = s
	}
	var lo, hi int
	switch {
	case rng == "*":
		lo, hi = min, max
	case strings.ContainsRune(rng, '-'):
		a, b, _ := strings.Cut(rng, "-")
		var e1, e2 error
		lo, e1 = strconv.Atoi(a)
		hi, e2 = strconv.Atoi(b)
		if e1 != nil || e2 != nil {
			return 0, fmt.Errorf("invalid range %q", part)
		}
	default:
		n, err := strconv.Atoi(rng)
		if err != nil {
			return 0, fmt.Errorf("invalid value %q", part)
		}
		lo = n
		if hasStep {
			hi = max
		} else {
			hi = n
		}
	}
	if lo < min || hi > max || lo > hi {
		return 0, fmt.Errorf("term %q out of range [%d,%d]", part, min, max)
	}
	var mask uint64
	for v := lo; v <= hi; v += step {
		mask |= 1 << uint(v)
	}
	return mask, nil
}

func fullMask(min, max int) uint64 {
	var m uint64
	for v := min; v <= max; v++ {
		m |= 1 << uint(v)
	}
	return m
}

func bitSet(mask uint64, v int) bool {
	if v < 0 || v > 63 {
		return false
	}
	return mask&(1<<uint(v)) != 0
}

// Next returns the next activation strictly after t.
func (c *CronSchedule) Next(t time.Time) time.Time {
	t = t.Truncate(time.Minute).Add(time.Minute)
	limit := t.AddDate(5, 0, 0)
	for t.Before(limit) {
		if !bitSet(c.month, int(t.Month())) {
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, 0)
			continue
		}
		if !c.dayMatches(t) {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, 1)
			continue
		}
		if !bitSet(c.hour, t.Hour()) {
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location()).Add(time.Hour)
			continue
		}
		if !bitSet(c.minute, t.Minute()) {
			t = t.Add(time.Minute)
			continue
		}
		return t
	}
	return time.Time{}
}

// dayMatches implements the standard cron rule: when both day-of-month and
// day-of-week are restricted, a day matches if EITHER matches; otherwise both.
func (c *CronSchedule) dayMatches(t time.Time) bool {
	domHit := bitSet(c.dom, t.Day())
	dowHit := bitSet(c.dow, int(t.Weekday()))
	if !c.domStar && !c.dowStar {
		return domHit || dowHit
	}
	return domHit && dowHit
}
