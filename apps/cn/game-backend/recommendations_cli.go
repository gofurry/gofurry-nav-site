package main

import (
	"context"
	"fmt"
	"time"

	v2service "github.com/gofurry/gofurry-game-backend/apps/game/v2/service"
	"github.com/gofurry/gofurry-game-backend/roof/env"
	"github.com/spf13/cobra"
)

func newRecommendationsCommand(options *cliOptions) *cobra.Command {
	root := &cobra.Command{Use: "recommendations", Short: "Maintain the private recommendation cache", Args: cobra.NoArgs}
	rebuild := &cobra.Command{Use: "rebuild", Short: "Recompute all eligible source games (top 64 each)", Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, _ []string) error { return env.LoadServerConfig(options.configFile) },
		RunE: func(cmd *cobra.Command, _ []string) error {
			gf := new(goFurry)
			if err := gf.initPostgres(); err != nil {
				return fmt.Errorf("connect to configured database failed")
			}
			defer gf.pool.Close()
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Minute)
			defer cancel()
			result, err := v2service.NewReadModelServiceWithReader(gf.readDAO).RebuildRecommendations(ctx)
			fmt.Fprintf(cmd.OutOrStdout(), "algorithm=similar-v2.4.0-hybrid-cbf total=%d rebuilt=%d failed=%d\n", result.Total, result.Rebuilt, result.Failed)
			return err
		}}
	root.AddCommand(rebuild)
	return root
}
