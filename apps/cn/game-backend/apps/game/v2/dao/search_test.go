package dao

import (
	"strings"
	"testing"
	"time"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

func TestBuildSearchWhereSeparatesAvailableAndPlannedWindows(t *testing.T) {
	start := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC)

	availableWhere, _ := buildSearchWhere(v2models.GameV2SearchPageQuery{
		Availability:     "available",
		PubStartTime:     start,
		PubEndTime:       end,
		PlannedStartTime: start,
		PlannedEndTime:   end,
	})
	if !strings.Contains(availableWhere, "rs.availability") || !strings.Contains(availableWhere, "fa.window_start") {
		t.Fatalf("available where = %s", availableWhere)
	}
	if strings.Contains(availableWhere, "rs.window_start") {
		t.Fatalf("available search must not apply planned window: %s", availableWhere)
	}

	upcomingWhere, _ := buildSearchWhere(v2models.GameV2SearchPageQuery{
		Availability:     "upcoming",
		PubStartTime:     start,
		PubEndTime:       end,
		PlannedStartTime: start,
		PlannedEndTime:   end,
	})
	if !strings.Contains(upcomingWhere, "rs.availability") || !strings.Contains(upcomingWhere, "rs.window_start") {
		t.Fatalf("upcoming where = %s", upcomingWhere)
	}
	if strings.Contains(upcomingWhere, "fa.window_start") {
		t.Fatalf("upcoming search must not apply first-available window: %s", upcomingWhere)
	}
}

func TestSearchOrderUpcomingUsesCanonicalPlannedWindow(t *testing.T) {
	order := searchOrder(v2models.GameV2SearchPageQuery{Availability: "upcoming", TimeOrder: true})
	parts := []string{
		"rs.window_end >= CURRENT_DATE",
		"rs.window_end ASC NULLS LAST",
		"rs.window_start ASC NULLS LAST",
		"CASE rs.precision",
	}
	previous := -1
	for _, part := range parts {
		index := strings.Index(order, part)
		if index <= previous {
			t.Fatalf("planned order %q missing or out of order in %q", part, order)
		}
		previous = index
	}
	if strings.Contains(order, "g.create_time DESC") {
		t.Fatalf("upcoming order must not use collection time as its primary schedule: %s", order)
	}
}
