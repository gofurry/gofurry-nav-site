package backfill

import "testing"

func TestParseLegacyRelease(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		precision string
		start     string
		end       string
	}{
		"2009.11.16":       {"day", "2009-11-16", "2009-11-16"},
		"2009-11-16":       {"day", "2009-11-16", "2009-11-16"},
		"2009年11月16日":      {"day", "2009-11-16", "2009-11-16"},
		"2009 年 11 月 16 日": {"day", "2009-11-16", "2009-11-16"},
		"16 Nov, 2009":     {"day", "2009-11-16", "2009-11-16"},
		"Nov 2009":         {"month", "2009-11-01", "2009-11-30"},
		"Q4 2009":          {"quarter", "2009-10-01", "2009-12-31"},
		"2009":             {"year", "2009-01-01", "2009-12-31"},
	}
	for input, want := range tests {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			got, ok := parseLegacyRelease(input)
			if !ok {
				t.Fatal("expected a parsed release")
			}
			if got.Precision != want.precision || got.WindowStart.Format("2006-01-02") != want.start || got.WindowEnd.Format("2006-01-02") != want.end {
				t.Fatalf("unexpected release: %#v", got)
			}
		})
	}
}

func TestParseLegacyReleaseRejectsAmbiguousAndInvalidValues(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"soon", "即将推出", "2026-02-30", "Spring 2026", ""} {
		if _, ok := parseLegacyRelease(input); ok {
			t.Errorf("expected %q to be rejected", input)
		}
	}
}
