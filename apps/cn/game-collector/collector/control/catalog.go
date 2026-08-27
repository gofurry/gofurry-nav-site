package control

import (
	"context"
	"fmt"
	"time"

	gamesqlc "github.com/gofurry/gofurry-game-collector/internal/db/game/sqlc"
	"github.com/gofurry/gofurry-game-collector/roof/env"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	JobKeyMetadata = "game.metadata"
	JobKeyPlayers  = "game.players"

	PriorityScheduled = int32(100)
	PriorityCatchUp   = int32(150)
	PriorityManual    = int32(200)
	PriorityEntity    = int32(300)

	ConcurrencySteam = "steam"
)

type Capability struct {
	JobKey      string
	Tasks       []string
	PointInTime bool
}

var capabilities = map[string]Capability{
	JobKeyMetadata: {JobKey: JobKeyMetadata, Tasks: []string{"details", "news"}},
	JobKeyPlayers:  {JobKey: JobKeyPlayers, Tasks: []string{"players"}, PointInTime: true},
}

func capability(jobKey string) (Capability, bool) {
	item, ok := capabilities[jobKey]
	return item, ok
}

func ensureSchedules(ctx context.Context, queries *gamesqlc.Queries) error {
	playerHours := env.GetServerConfig().Collector.Game.GamePlayerInterval
	if playerHours <= 0 {
		playerHours = 1
	}
	playerSeconds := int64((time.Duration(playerHours) * time.Hour) / time.Second)
	cronExpression := "0 3 * * *"
	if err := queries.EnsureGameCollectionSchedule(ctx, gamesqlc.EnsureGameCollectionScheduleParams{
		JobKey:              JobKeyMetadata,
		Name:                "Game metadata refresh",
		ScheduleKind:        "cron",
		CronExpression:      &cronExpression,
		Timezone:            "Asia/Shanghai",
		MisfirePolicy:       "catch_up_once",
		MisfireGraceSeconds: 300,
		Priority:            PriorityScheduled,
		ConcurrencyKey:      ConcurrencySteam,
	}); err != nil {
		return fmt.Errorf("ensure %s schedule: %w", JobKeyMetadata, err)
	}
	anchor := pgtype.Timestamptz{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}
	if err := queries.EnsureGameCollectionSchedule(ctx, gamesqlc.EnsureGameCollectionScheduleParams{
		JobKey:              JobKeyPlayers,
		Name:                "Game player-count samples",
		ScheduleKind:        "interval",
		IntervalSeconds:     &playerSeconds,
		AnchorAt:            anchor,
		Timezone:            "UTC",
		MisfirePolicy:       "skip",
		MisfireGraceSeconds: 90,
		Priority:            PriorityScheduled,
		ConcurrencyKey:      ConcurrencySteam,
	}); err != nil {
		return fmt.Errorf("ensure %s schedule: %w", JobKeyPlayers, err)
	}
	return nil
}
