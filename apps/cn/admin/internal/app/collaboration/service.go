package collaboration

import (
	"context"
	"errors"
	"net/http"

	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	gamesqlc "github.com/gofurry/gofurry-admin/internal/db/game/sqlc"
	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const conflictMessage = "此内容刚刚被其他成员修改，请重新加载后重试。"

// Only GFA exposes a pool to mutation paths. Game/Nav dependencies expose their
// generated read queries; formal content creation remains in its owning API.
type Service struct {
	admin   *pgxpool.Pool
	queries *adminsqlc.Queries
	game    *gamesqlc.Queries
	nav     *navsqlc.Queries
	audit   *audit.Logger
}

func New(admin, game, nav *pgxpool.Pool, logger *audit.Logger) *Service {
	return &Service{admin: admin, queries: adminsqlc.New(admin), game: gamesqlc.New(game), nav: navsqlc.New(nav), audit: logger}
}
func databaseError(err error) common.Error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return common.NewError(common.RETURN_FAILED, http.StatusNotFound, "内容不存在")
	}
	return common.NewDaoError("无法读写协作数据")
}
func versionError(err error) common.Error {
	if errors.Is(err, pgx.ErrNoRows) {
		return common.NewConflictError(conflictMessage)
	}
	return databaseError(err)
}
func actor(meta audit.Meta) (int64, common.Error) {
	if meta.OperatorAccountID == nil || *meta.OperatorAccountID <= 0 {
		return 0, common.NewError(common.RETURN_FAILED, http.StatusUnauthorized, "登录后才能修改协作内容")
	}
	return *meta.OperatorAccountID, nil
}
func (s *Service) Get(ctx context.Context, id int64) (Idea, common.Error) {
	row, err := s.queries.GetContentIdea(ctx, id)
	return row, databaseError(err)
}

// Deleting an idea never touches its linked formal resource. The audit snapshot
// retains the removed GFA record and commits atomically with the deletion.
func (s *Service) Delete(ctx context.Context, meta audit.Meta, id, version int64) common.Error {
	if _, e := actor(meta); e != nil {
		return e
	}
	if version <= 0 {
		return common.NewValidationError("version is required")
	}
	tx, err := s.admin.Begin(ctx)
	if err != nil {
		return databaseError(err)
	}
	defer tx.Rollback(ctx)
	before, err := s.queries.WithTx(tx).DeleteContentIdeaVersioned(ctx, adminsqlc.DeleteContentIdeaVersionedParams{ID: id, Version: version})
	if err != nil {
		return versionError(err)
	}
	if e := s.audit.LogTx(ctx, tx, meta, "collaboration.idea.delete", "collaboration.idea", id, before, nil); e != nil {
		return e
	}
	return databaseError(tx.Commit(ctx))
}
func (s *Service) Summary(ctx context.Context) (Summary, error) {
	return s.queries.CountContentIdeaSummary(ctx)
}
func (s *Service) List(ctx context.Context, f Filters, accountID int64) (IdeaPage, common.Error) {
	if f.Status == "" {
		f.Status = "active"
	}
	if f.Researcher == "" {
		f.Researcher = "all"
	}
	if !oneOf(f.Status, "active", "all", "idea", "researching", "landed", "shelved") || !oneOf(f.Kind, "", "game", "site", "other") || !oneOf(f.Priority, "", "normal", "high") || !oneOf(f.Researcher, "all", "me", "unassigned") {
		return IdeaPage{}, common.NewValidationError("invalid idea filters")
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Page > 1000000 {
		return IdeaPage{}, common.NewValidationError("page too large")
	}
	if f.PageSize < 1 {
		f.PageSize = 50
	}
	if f.PageSize > 100 {
		f.PageSize = 100
	}
	count, err := s.queries.CountContentIdeas(ctx, adminsqlc.CountContentIdeasParams{Kind: f.Kind, Status: f.Status, Priority: f.Priority, Researcher: f.Researcher, AccountID: accountID, Keyword: f.Keyword})
	if err != nil {
		return IdeaPage{}, databaseError(err)
	}
	rows, err := s.queries.ListContentIdeas(ctx, adminsqlc.ListContentIdeasParams{Kind: f.Kind, Status: f.Status, Priority: f.Priority, Researcher: f.Researcher, AccountID: accountID, Keyword: f.Keyword, RowLimit: int32(f.PageSize), RowOffset: int32((f.Page - 1) * f.PageSize)})
	return IdeaPage{Total: count, List: rows}, databaseError(err)
}
func (s *Service) Create(ctx context.Context, meta audit.Meta, input Input) (Idea, common.Error) {
	account, e := actor(meta)
	if e != nil {
		return Idea{}, e
	}
	in, key, e := normalize(input)
	if e != nil {
		return Idea{}, e
	}
	tx, err := s.admin.Begin(ctx)
	if err != nil {
		return Idea{}, databaseError(err)
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	row, err := q.InsertContentIdea(ctx, adminsqlc.InsertContentIdeaParams{Kind: in.Kind, Title: nullable(in.Title), Source: nullable(in.Source), SourceKey: key, Note: in.Note, Priority: in.Priority, CreatedByAccountID: account})
	if err != nil {
		return Idea{}, databaseError(err)
	}
	if e := s.audit.LogTx(ctx, tx, meta, "collaboration.idea.create", "collaboration.idea", row.ID, nil, row); e != nil {
		return Idea{}, e
	}
	result, err := q.GetContentIdea(ctx, row.ID)
	if err != nil {
		return Idea{}, databaseError(err)
	}
	return result, databaseError(tx.Commit(ctx))
}
func (s *Service) Update(ctx context.Context, meta audit.Meta, id int64, input Input) (Idea, common.Error) {
	if _, e := actor(meta); e != nil {
		return Idea{}, e
	}
	if input.Version <= 0 {
		return Idea{}, common.NewValidationError("version is required")
	}
	in, key, e := normalize(input)
	if e != nil {
		return Idea{}, e
	}
	tx, err := s.admin.Begin(ctx)
	if err != nil {
		return Idea{}, databaseError(err)
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	before, err := q.GetContentIdea(ctx, id)
	if err != nil {
		return Idea{}, versionError(err)
	}
	if before.Version != in.Version {
		return Idea{}, common.NewConflictError(conflictMessage)
	}
	if before.Status == "landed" && before.Kind != in.Kind {
		return Idea{}, common.NewConflictError("请先重新打开已落地的内容，再修改类型")
	}
	row, err := q.UpdateContentIdeaVersioned(ctx, adminsqlc.UpdateContentIdeaVersionedParams{ID: id, Version: in.Version, Kind: in.Kind, Title: nullable(in.Title), Source: nullable(in.Source), SourceKey: key, Note: in.Note, Priority: in.Priority})
	if err != nil {
		return Idea{}, versionError(err)
	}
	if e := s.audit.LogTx(ctx, tx, meta, "collaboration.idea.update", "collaboration.idea", id, before, row); e != nil {
		return Idea{}, e
	}
	result, err := q.GetContentIdea(ctx, id)
	if err != nil {
		return Idea{}, databaseError(err)
	}
	return result, databaseError(tx.Commit(ctx))
}

func (s *Service) Transition(ctx context.Context, meta audit.Meta, id int64, action string, in Transition) (Idea, common.Error) {
	account, e := actor(meta)
	if e != nil {
		return Idea{}, e
	}
	if in.Version <= 0 {
		return Idea{}, common.NewValidationError("version is required")
	}
	// Existence checks are explicit independent reads, not part of a cross-DB transaction.
	if action == "link" {
		before, e := s.Get(ctx, id)
		if e != nil {
			return Idea{}, e
		}
		if !oneOf(in.Kind, "game", "site") || in.ResourceID <= 0 || (before.Kind != "other" && before.Kind != in.Kind) {
			return Idea{}, common.NewValidationError("关联内容类型或 ID 无效")
		}
		var err error
		if in.Kind == "game" {
			_, err = s.game.GetGameForCollaborationLink(ctx, in.ResourceID)
		} else {
			_, err = s.nav.GetSiteForCollaborationLink(ctx, in.ResourceID)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return Idea{}, common.NewValidationError("正式内容不存在")
		}
		if err != nil {
			return Idea{}, databaseError(err)
		}
	}
	tx, err := s.admin.Begin(ctx)
	if err != nil {
		return Idea{}, databaseError(err)
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	before, err := q.GetContentIdea(ctx, id)
	if err != nil {
		return Idea{}, versionError(err)
	}
	if before.Version != in.Version {
		return Idea{}, common.NewConflictError(conflictMessage)
	}
	if action == "research" && before.Status == "researching" {
		if before.ResearchingByAccountID != nil && *before.ResearchingByAccountID == account {
			return before, nil
		}
		return Idea{}, common.NewConflictError("该内容目前由 " + before.ResearcherName + " 整理")
	}
	var row adminsqlc.GfaContentIdea
	switch action {
	case "research":
		row, err = q.ResearchContentIdeaVersioned(ctx, adminsqlc.ResearchContentIdeaVersionedParams{ID: id, Version: in.Version, AccountID: &account})
	case "release":
		row, err = q.ReleaseContentIdeaVersioned(ctx, adminsqlc.ReleaseContentIdeaVersionedParams{ID: id, Version: in.Version})
	case "shelve":
		row, err = q.ShelveContentIdeaVersioned(ctx, adminsqlc.ShelveContentIdeaVersionedParams{ID: id, Version: in.Version})
	case "restore":
		row, err = q.RestoreContentIdeaVersioned(ctx, adminsqlc.RestoreContentIdeaVersionedParams{ID: id, Version: in.Version})
	case "link":
		if before.Kind != "other" && before.Kind != in.Kind {
			return Idea{}, common.NewConflictError(conflictMessage)
		}
		row, err = q.LinkContentIdeaVersioned(ctx, adminsqlc.LinkContentIdeaVersionedParams{ID: id, Version: in.Version, LinkedKind: &in.Kind, LinkedResourceID: &in.ResourceID})
	case "land":
		row, err = q.LandOtherContentIdeaVersioned(ctx, adminsqlc.LandOtherContentIdeaVersionedParams{ID: id, Version: in.Version})
	case "reopen":
		row, err = q.ReopenContentIdeaVersioned(ctx, adminsqlc.ReopenContentIdeaVersionedParams{ID: id, Version: in.Version})
	default:
		return Idea{}, common.NewValidationError("unknown transition")
	}
	if err != nil {
		return Idea{}, versionError(err)
	}
	if e := s.audit.LogTx(ctx, tx, meta, "collaboration.idea."+action, "collaboration.idea", id, before, row); e != nil {
		return Idea{}, e
	}
	result, err := q.GetContentIdea(ctx, id)
	if err != nil {
		return Idea{}, databaseError(err)
	}
	return result, databaseError(tx.Commit(ctx))
}
