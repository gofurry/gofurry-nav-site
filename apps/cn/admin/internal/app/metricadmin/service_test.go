package metricadmin

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestRatesUseKnownAndEligibleDenominators(t *testing.T) {
	adoption, coverage := rates(3, 1, 8)
	if adoption == nil || *adoption != 0.75 {
		t.Fatalf("adoption=%v, want 0.75", adoption)
	}
	if coverage == nil || *coverage != 0.5 {
		t.Fatalf("coverage=%v, want 0.5", coverage)
	}

	adoption, coverage = rates(0, 0, 0)
	if adoption != nil || coverage != nil {
		t.Fatalf("zero denominators must produce null rates: adoption=%v coverage=%v", adoption, coverage)
	}
}

func TestLagDaysUsesNextOrderedDay(t *testing.T) {
	date := func(value string) pgtype.Date {
		parsed, err := time.Parse(time.DateOnly, value)
		if err != nil {
			t.Fatal(err)
		}
		return pgtype.Date{Time: parsed, Valid: true}
	}

	lag := lagDays(date("2026-08-01"), date("2026-08-03"), date("2026-08-05"))
	if lag == nil || *lag != 2 {
		t.Fatalf("lag=%v, want 2", lag)
	}
	if got := lagDays(pgtype.Date{}, pgtype.Date{}, date("2026-08-05")); got != nil {
		t.Fatalf("missing source start must not invent lag: %v", got)
	}
}
