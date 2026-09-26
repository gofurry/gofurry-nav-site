package collaboration

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

func (s *Service) Preview(ctx context.Context, items []Input) ([]PreviewRow, common.Error) {
	if len(items) == 0 || len(items) > 500 {
		return nil, common.NewValidationError("每批需要 1–500 条内容")
	}
	rows := make([]PreviewRow, len(items))
	keys, hosts, appids := []string{}, []string{}, []int64{}
	seen := map[string]int{}
	for index, input := range items {
		in, key, e := normalize(input)
		row := PreviewRow{Input: in, Index: index, SourceKey: key, Valid: e == nil, Errors: []string{}, Warnings: []string{}}
		if e != nil {
			row.Errors = append(row.Errors, e.GetMsg())
		}
		if key != nil && row.Valid {
			row.DisplayHint = *key
			if previous, ok := seen[*key]; ok {
				row.Warnings = append(row.Warnings, "batch_duplicate")
				row.BatchDuplicateOf = ptr(previous)
			} else {
				seen[*key] = index
				keys = append(keys, *key)
			}
			if strings.HasPrefix(*key, "steam:") {
				id, _ := strconv.ParseInt(strings.TrimPrefix(*key, "steam:"), 10, 64)
				appids = append(appids, id)
			}
			if strings.HasPrefix(*key, "host:") {
				hosts = append(hosts, strings.TrimPrefix(*key, "host:"))
			}
		}
		rows[index] = row
	}
	ideaMatches, resourceMatches := map[string]Match{}, map[string]Match{}
	if len(keys) > 0 {
		existing, err := s.queries.ListContentIdeasBySourceKeys(ctx, keys)
		if err != nil {
			return nil, databaseError(err)
		}
		for _, row := range existing {
			ideaMatches[text(row.SourceKey)] = Match{ID: row.ID, Kind: row.Kind, Title: text(row.Title)}
		}
	}
	if len(appids) > 0 {
		existing, err := s.game.ListGamesByAppIDsForCollaboration(ctx, appids)
		if err != nil {
			return nil, databaseError(err)
		}
		for _, row := range existing {
			key := "steam:" + strconv.FormatInt(row.Appid, 10)
			if _, ok := resourceMatches[key]; !ok {
				resourceMatches[key] = Match{ID: row.ID, Kind: "game", Title: row.Name}
			}
		}
	}
	if len(hosts) > 0 {
		existing, err := s.nav.ListSitesByNormalizedHostsForCollaboration(ctx, hosts)
		if err != nil {
			return nil, databaseError(err)
		}
		for _, row := range existing {
			key := "host:" + row.Host
			if _, ok := resourceMatches[key]; !ok {
				resourceMatches[key] = Match{ID: row.ID, Kind: "site", Title: row.Name}
			}
		}
	}
	for i := range rows {
		if !rows[i].Valid {
			continue
		}
		key := text(rows[i].SourceKey)
		if match, ok := ideaMatches[key]; ok {
			rows[i].IdeaMatch = &match
			rows[i].Warnings = append(rows[i].Warnings, "idea_duplicate")
		}
		if match, ok := resourceMatches[key]; ok {
			rows[i].ResourceMatch = &match
			rows[i].Warnings = append(rows[i].Warnings, "resource_exists")
		}
	}
	return rows, nil
}

func (s *Service) Batch(ctx context.Context, meta audit.Meta, input BatchInput) (BatchResult, common.Error) {
	account, e := actor(meta)
	if e != nil {
		return BatchResult{}, e
	}
	// Recompute from original candidates. The client preview is never trusted.
	rows, e := s.Preview(ctx, input.Items)
	if e != nil {
		return BatchResult{}, e
	}
	params := []adminsqlc.InsertContentIdeasBatchParams{}
	skipKnown := input.SkipKnown == nil || *input.SkipKnown
	kinds := map[string]bool{}
	for _, row := range rows {
		if !row.Valid || (skipKnown && len(row.Warnings) > 0) {
			continue
		}
		kinds[row.Kind] = true
		params = append(params, adminsqlc.InsertContentIdeasBatchParams{Kind: row.Kind, Title: nullable(row.Title), Source: nullable(row.Source), SourceKey: row.SourceKey, Note: row.Note, Priority: row.Priority, CreatedByAccountID: account})
	}
	result := BatchResult{SkippedCount: len(rows) - len(params)}
	if len(params) == 0 {
		return result, nil
	}
	tx, err := s.admin.Begin(ctx)
	if err != nil {
		return result, databaseError(err)
	}
	defer tx.Rollback(ctx)
	result.InsertedCount, err = s.queries.WithTx(tx).InsertContentIdeasBatch(ctx, params)
	if err != nil {
		return BatchResult{}, databaseError(err)
	}
	kind := "mixed"
	if len(kinds) == 1 {
		for value := range kinds {
			kind = value
		}
	}
	snapshot := struct {
		Kind string `json:"kind"`
		BatchResult
	}{kind, result}
	if e := s.audit.LogTx(ctx, tx, meta, "collaboration.idea.batch_create", "collaboration.idea", nil, nil, snapshot); e != nil {
		return BatchResult{}, e
	}
	return result, databaseError(tx.Commit(ctx))
}
