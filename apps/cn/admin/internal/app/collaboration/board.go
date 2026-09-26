package collaboration

import (
	"context"
	"errors"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"
	"github.com/gofurry/gofurry-admin/pkg/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func boardError(err error) common.Error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && (pg.Code == "23503" || pg.Code == "23505") {
		return common.NewConflictError("连接端点已变化或连接已存在，请重新加载。")
	}
	return versionError(err)
}
func validateGeometry(in BoardGeometry) common.Error {
	if in.X < -100000 || in.Y < -100000 || in.X > 100000 || in.Y > 100000 || in.Width < 48 || in.Width > 1600 || in.Height < 40 || in.Height > 1600 || in.ZIndex < 0 || in.ZIndex > 1000000 {
		return common.NewValidationError("画布位置或尺寸超出范围")
	}
	return nil
}
func normalizeBoardNode(in *BoardNodeInput) common.Error {
	in.Title, in.Body = strings.TrimSpace(in.Title), strings.TrimSpace(in.Body)
	if in.Kind == "" {
		in.Kind = "note"
	}
	if in.Color == "" {
		in.Color = "sand"
	}
	if !oneOf(in.Kind, "note", "card", "text", "rectangle", "ellipse", "arrow") || !oneOf(in.Color, "sand", "blue", "green", "rose", "slate") || (in.Rotation != 0 && in.Rotation != 90 && in.Rotation != 180 && in.Rotation != 270) {
		return common.NewValidationError("画布元素类型或样式无效")
	}
	if strings.ContainsRune(in.Title+in.Body, 0) || utf8.RuneCountInString(in.Title) > 200 || utf8.RuneCountInString(in.Body) > 10000 {
		return common.NewValidationError("标题最多 200 字，正文最多 10000 字")
	}
	if in.ReferenceKind != "" || in.ReferenceID != 0 {
		if in.Kind != "card" || !oneOf(in.ReferenceKind, "idea", "game", "site") || in.ReferenceID <= 0 {
			return common.NewValidationError("内容卡片引用无效")
		}
	}
	if !oneOf(in.Kind, "rectangle", "ellipse", "arrow") && in.Title == "" && in.Body == "" && in.ReferenceID == 0 {
		return common.NewValidationError("请填写标题或正文")
	}
	return validateGeometry(in.BoardGeometry)
}
func (s *Service) validateBoardReference(ctx context.Context, in BoardNodeInput) common.Error {
	var err error
	switch in.ReferenceKind {
	case "idea":
		_, err = s.queries.GetContentIdea(ctx, in.ReferenceID)
	case "game":
		_, err = s.game.GetGameForCollaborationLink(ctx, in.ReferenceID)
	case "site":
		_, err = s.nav.GetSiteForCollaborationLink(ctx, in.ReferenceID)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return common.NewValidationError("引用内容不存在或已被删除")
	}
	return databaseError(err)
}
func (s *Service) Board(ctx context.Context) (BoardDocument, common.Error) {
	result := BoardDocument{Nodes: []BoardNode{}, Edges: []BoardEdge{}, References: []BoardReference{}}
	tx, err := s.admin.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return result, databaseError(err)
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	result.Nodes, err = q.ListBoardNodes(ctx)
	if err != nil {
		return result, databaseError(err)
	}
	result.Edges, err = q.ListBoardEdges(ctx)
	if err != nil {
		return result, databaseError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return result, databaseError(err)
	}
	// References are read projections, not cross-database foreign keys/transactions.
	result.References, err = s.boardReferences(ctx, result.Nodes)
	return result, databaseError(err)
}
func (s *Service) SaveBoardNode(ctx context.Context, meta audit.Meta, id int64, in BoardNodeInput) (BoardNode, common.Error) {
	account, e := actor(meta)
	if e != nil {
		return BoardNode{}, e
	}
	if e = normalizeBoardNode(&in); e != nil {
		return BoardNode{}, e
	}
	if id > 0 && in.Version <= 0 {
		return BoardNode{}, common.NewValidationError("version is required")
	}
	if e = s.validateBoardReference(ctx, in); e != nil {
		return BoardNode{}, e
	}
	tx, err := s.admin.Begin(ctx)
	if err != nil {
		return BoardNode{}, databaseError(err)
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	var before, after BoardNode
	action := "collaboration.board_node.create"
	var referenceID *int64
	if in.ReferenceID > 0 {
		referenceID = &in.ReferenceID
	}
	if id == 0 {
		after, err = q.InsertBoardNode(ctx, adminsqlc.InsertBoardNodeParams{Kind: in.Kind, Title: in.Title, Body: in.Body, Color: in.Color, Rotation: in.Rotation, ReferenceKind: nullable(in.ReferenceKind), ReferenceID: referenceID, X: in.X, Y: in.Y, Width: in.Width, Height: in.Height, ZIndex: in.ZIndex, AccountID: account})
	} else {
		before, err = q.GetBoardNodeForUpdate(ctx, id)
		if err != nil {
			return BoardNode{}, versionError(err)
		}
		if before.Kind != in.Kind {
			return BoardNode{}, common.NewValidationError("不能改变已创建元素的类型")
		}
		after, err = q.UpdateBoardNodeVersioned(ctx, adminsqlc.UpdateBoardNodeVersionedParams{ID: id, Version: in.Version, Title: in.Title, Body: in.Body, Color: in.Color, Rotation: in.Rotation, ReferenceKind: nullable(in.ReferenceKind), ReferenceID: referenceID, UpdatedByAccountID: account})
		action = "collaboration.board_node.update"
	}
	if err != nil {
		return BoardNode{}, boardError(err)
	}
	var snapshot any
	if id > 0 {
		snapshot = before
	}
	if e = s.audit.LogTx(ctx, tx, meta, action, "collaboration.board_node", after.ID, snapshot, after); e != nil {
		return BoardNode{}, e
	}
	return after, databaseError(tx.Commit(ctx))
}

// A multi-selection move is one transaction. Any stale member rolls back the
// entire gesture. This endpoint accepts geometry only and intentionally omits Audit.
func (s *Service) MoveBoardNodes(ctx context.Context, meta audit.Meta, in []BoardLayout) ([]BoardNode, common.Error) {
	account, e := actor(meta)
	if e != nil {
		return nil, e
	}
	if len(in) == 0 || len(in) > 500 {
		return nil, common.NewValidationError("每次移动需要 1–500 个元素")
	}
	seen := map[int64]bool{}
	for _, item := range in {
		if item.ID <= 0 || item.Version <= 0 || seen[item.ID] {
			return nil, common.NewValidationError("元素 ID/version 无效或重复")
		}
		seen[item.ID] = true
		if e = validateGeometry(item.BoardGeometry); e != nil {
			return nil, e
		}
	}
	items := append([]BoardLayout(nil), in...)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	tx, err := s.admin.Begin(ctx)
	if err != nil {
		return nil, databaseError(err)
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	result := make([]BoardNode, 0, len(items))
	for _, item := range items {
		node, err := q.MoveBoardNodeVersioned(ctx, adminsqlc.MoveBoardNodeVersionedParams{ID: item.ID, Version: item.Version, X: item.X, Y: item.Y, Width: item.Width, Height: item.Height, ZIndex: item.ZIndex, UpdatedByAccountID: account})
		if err != nil {
			return nil, versionError(err)
		}
		result = append(result, node)
	}
	return result, databaseError(tx.Commit(ctx))
}
func (s *Service) DeleteBoardNode(ctx context.Context, meta audit.Meta, id, version int64) common.Error {
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
	q := s.queries.WithTx(tx)
	before, err := q.GetBoardNodeForUpdate(ctx, id)
	if err != nil {
		return versionError(err)
	}
	if before.Version != version {
		return common.NewConflictError(conflictMessage)
	}
	edges, err := q.ListIncidentBoardEdgesForUpdate(ctx, id)
	if err != nil {
		return databaseError(err)
	}
	if _, err = q.DeleteBoardNodeVersioned(ctx, adminsqlc.DeleteBoardNodeVersionedParams{ID: id, Version: version}); err != nil {
		return versionError(err)
	}
	snapshot := struct {
		Node  BoardNode   `json:"node"`
		Edges []BoardEdge `json:"edges"`
	}{before, edges}
	if e := s.audit.LogTx(ctx, tx, meta, "collaboration.board_node.delete", "collaboration.board_node", id, snapshot, nil); e != nil {
		return e
	}
	return databaseError(tx.Commit(ctx))
}
func normalizeBoardEdge(in *BoardEdgeInput) common.Error {
	in.Label = strings.TrimSpace(in.Label)
	if in.Routing == "" {
		in.Routing = "curve"
	}
	if in.Color == "" {
		in.Color = "slate"
	}
	if in.SourceID <= 0 || in.TargetID <= 0 || in.SourceID == in.TargetID || !oneOf(in.SourceHandle, "top", "right", "bottom", "left") || !oneOf(in.TargetHandle, "top", "right", "bottom", "left") || !oneOf(in.Routing, "curve", "step") || !oneOf(in.Color, "sand", "blue", "green", "rose", "slate") || strings.ContainsRune(in.Label, 0) || utf8.RuneCountInString(in.Label) > 200 {
		return common.NewValidationError("连线端点、样式或标签无效")
	}
	return nil
}
func (s *Service) SaveBoardEdge(ctx context.Context, meta audit.Meta, id int64, in BoardEdgeInput) (BoardEdge, common.Error) {
	account, e := actor(meta)
	if e != nil {
		return BoardEdge{}, e
	}
	if e = normalizeBoardEdge(&in); e != nil {
		return BoardEdge{}, e
	}
	if id > 0 && in.Version <= 0 {
		return BoardEdge{}, common.NewValidationError("version is required")
	}
	tx, err := s.admin.Begin(ctx)
	if err != nil {
		return BoardEdge{}, databaseError(err)
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	var before, after BoardEdge
	action := "collaboration.board_edge.create"
	if id == 0 {
		ids, lockErr := q.LockBoardEndpoints(ctx, []int64{in.SourceID, in.TargetID})
		if lockErr != nil {
			return BoardEdge{}, databaseError(lockErr)
		}
		if len(ids) != 2 {
			return BoardEdge{}, common.NewConflictError("连接端点不存在或不支持连线，请重新加载。")
		}
		after, err = q.InsertBoardEdge(ctx, adminsqlc.InsertBoardEdgeParams{SourceID: in.SourceID, TargetID: in.TargetID, SourceHandle: in.SourceHandle, TargetHandle: in.TargetHandle, Routing: in.Routing, Label: in.Label, Color: in.Color, Arrow: in.Arrow, AccountID: account})
	} else {
		before, err = q.GetBoardEdgeForUpdate(ctx, id)
		if err != nil {
			return BoardEdge{}, versionError(err)
		}
		if before.SourceID != in.SourceID || before.TargetID != in.TargetID || before.SourceHandle != in.SourceHandle || before.TargetHandle != in.TargetHandle {
			return BoardEdge{}, common.NewValidationError("修改端点请删除连线后重新连接")
		}
		after, err = q.UpdateBoardEdgeVersioned(ctx, adminsqlc.UpdateBoardEdgeVersionedParams{ID: id, Version: in.Version, Routing: in.Routing, Label: in.Label, Color: in.Color, Arrow: in.Arrow, UpdatedByAccountID: account})
		action = "collaboration.board_edge.update"
	}
	if err != nil {
		return BoardEdge{}, boardError(err)
	}
	var snapshot any
	if id > 0 {
		snapshot = before
	}
	if e = s.audit.LogTx(ctx, tx, meta, action, "collaboration.board_edge", after.ID, snapshot, after); e != nil {
		return BoardEdge{}, e
	}
	return after, databaseError(tx.Commit(ctx))
}
func (s *Service) DeleteBoardEdge(ctx context.Context, meta audit.Meta, id, version int64) common.Error {
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
	before, err := s.queries.WithTx(tx).DeleteBoardEdgeVersioned(ctx, adminsqlc.DeleteBoardEdgeVersionedParams{ID: id, Version: version})
	if err != nil {
		return versionError(err)
	}
	if e := s.audit.LogTx(ctx, tx, meta, "collaboration.board_edge.delete", "collaboration.board_edge", id, before, nil); e != nil {
		return e
	}
	return databaseError(tx.Commit(ctx))
}
