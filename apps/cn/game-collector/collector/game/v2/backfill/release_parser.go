package backfill

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	legacyChineseDay = regexp.MustCompile(`^([0-9]{4})\s*年\s*([0-9]{1,2})\s*月\s*([0-9]{1,2})\s*日$`)
	legacyQuarter    = regexp.MustCompile(`(?i)^Q([1-4])\s+([0-9]{4})$`)
	legacyYear       = regexp.MustCompile(`^[0-9]{4}$`)
)

type legacyRelease struct {
	Precision   string
	ExactDate   *time.Time
	Year        int
	Month       *int
	Quarter     *int
	WindowStart time.Time
	WindowEnd   time.Time
}

func parseLegacyRelease(raw string) (legacyRelease, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return legacyRelease{}, false
	}

	if match := legacyChineseDay.FindStringSubmatch(value); match != nil {
		year, _ := strconv.Atoi(match[1])
		month, _ := strconv.Atoi(match[2])
		day, _ := strconv.Atoi(match[3])
		return exactLegacyDate(year, month, day)
	}

	if match := legacyQuarter.FindStringSubmatch(value); match != nil {
		quarter, _ := strconv.Atoi(match[1])
		year, _ := strconv.Atoi(match[2])
		start := time.Date(year, time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, -1)
		return legacyRelease{Precision: "quarter", Year: year, Quarter: &quarter, WindowStart: start, WindowEnd: end}, true
	}

	if legacyYear.MatchString(value) {
		year, _ := strconv.Atoi(value)
		if year < 1 {
			return legacyRelease{}, false
		}
		return legacyRelease{
			Precision:   "year",
			Year:        year,
			WindowStart: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			WindowEnd:   time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC),
		}, true
	}

	for _, layout := range []string{"2006.01.02", "2006-01-02", "2 Jan, 2006", "Jan 2, 2006", "2 January, 2006", "January 2, 2006"} {
		if parsed, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return exactLegacyDate(parsed.Year(), int(parsed.Month()), parsed.Day())
		}
	}

	for _, layout := range []string{"January 2006", "Jan 2006"} {
		if parsed, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			month := int(parsed.Month())
			start := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
			return legacyRelease{
				Precision:   "month",
				Year:        parsed.Year(),
				Month:       &month,
				WindowStart: start,
				WindowEnd:   start.AddDate(0, 1, -1),
			}, true
		}
	}

	return legacyRelease{}, false
}

func exactLegacyDate(year, month, day int) (legacyRelease, bool) {
	if year < 1 || month < 1 || month > 12 || day < 1 {
		return legacyRelease{}, false
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if date.Year() != year || int(date.Month()) != month || date.Day() != day {
		return legacyRelease{}, false
	}
	monthValue := month
	return legacyRelease{
		Precision:   "day",
		ExactDate:   &date,
		Year:        year,
		Month:       &monthValue,
		WindowStart: date,
		WindowEnd:   date,
	}, true
}
