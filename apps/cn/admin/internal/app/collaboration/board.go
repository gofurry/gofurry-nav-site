package collaboration

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
)

func validateBoard(in BoardInput) common.Error {
	if strings.TrimSpace(in.Body) == "" || strings.ContainsRune(in.Body, 0) || utf8.RuneCountInString(in.Body) > 10000 {
		return common.NewValidationError("便笺正文需要 1–10000 字")
	}
	if in.X < 0 || in.Y < 0 || in.X > 20000 || in.Y > 20000 || in.Width < 160 || in.Width > 1600 || in.Height < 120 || in.Height > 1600 || in.ZIndex < 0 || in.ZIndex > 1000000 {
		return common.NewValidationError("便笺位置或尺寸超出范围")
	}
	return nil
}
func (s *Service) Board(ctx context.Context) ([]BoardNote, common.Error) {
	rows, err := s.queries.ListBoardNotes(ctx)
	return rows, databaseError(err)
}
func (s *Service) SaveBoard(ctx context.Context, meta audit.Meta, id int64, in BoardInput) (BoardNote, common.Error) {
	account, e := actor(meta)
	if e != nil {
		return BoardNote{}, e
	}
	if e := validateBoard(in); e != nil {
		return BoardNote{}, e
	}
	if id > 0 && in.Version <= 0 {
		return BoardNote{}, common.NewValidationError("version is required")
	}
	tx, err := s.admin.Begin(ctx)
	if err != nil {
		return BoardNote{}, databaseError(err)
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	var before, after BoardNote
	action := "collaboration.board_note.create"
	if id == 0 {
		after, err = q.InsertBoardNote(ctx, adminsqlc.InsertBoardNoteParams{Body: in.Body, X: in.X, Y: in.Y, Width: in.Width, Height: in.Height, ZIndex: in.ZIndex, AccountID: account})
	} else {
		before, err = q.GetBoardNote(ctx, id)
		if err != nil {
			return BoardNote{}, versionError(err)
		}
		after, err = q.UpdateBoardNoteVersioned(ctx, adminsqlc.UpdateBoardNoteVersionedParams{ID: id, Version: in.Version, Body: in.Body, X: in.X, Y: in.Y, Width: in.Width, Height: in.Height, ZIndex: in.ZIndex, UpdatedByAccountID: account})
		action = "collaboration.board_note.update"
	}
	if err != nil {
		return BoardNote{}, versionError(err)
	}
	if id == 0 || before.Body != after.Body {
		var snapshot any
		if id > 0 {
			snapshot = before
		}
		if e := s.audit.LogTx(ctx, tx, meta, action, "collaboration.board_note", after.ID, snapshot, after); e != nil {
			return BoardNote{}, e
		}
	}
	return after, databaseError(tx.Commit(ctx))
}
func (s *Service) DeleteBoard(ctx context.Context, meta audit.Meta, id, version int64) common.Error {
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
	before, err := s.queries.WithTx(tx).DeleteBoardNoteVersioned(ctx, adminsqlc.DeleteBoardNoteVersionedParams{ID: id, Version: version})
	if err != nil {
		return versionError(err)
	}
	if e := s.audit.LogTx(ctx, tx, meta, "collaboration.board_note.delete", "collaboration.board_note", id, before, nil); e != nil {
		return e
	}
	return databaseError(tx.Commit(ctx))
}
