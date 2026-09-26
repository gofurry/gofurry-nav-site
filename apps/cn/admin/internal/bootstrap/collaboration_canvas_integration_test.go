package bootstrap_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofurry/gofurry-admin/internal/app/collaboration"
	"github.com/gofurry/gofurry-admin/internal/app/shared/audit"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testCollaborationCanvas(t *testing.T, ctx context.Context, admin *pgxpool.Pool, s *collaboration.Service, app *fiber.App, meta, other audit.Meta, ideaID, gameID, siteID int64, trace *collaborationTrace) {
	t.Helper()
	api := collaboration.NewAPI(s)
	app.Get("/board", api.Board)
	app.Post("/board/nodes", api.CreateBoardNode)
	app.Put("/board/nodes/layout", api.MoveBoardNodes)
	app.Put("/board/nodes/:id", api.UpdateBoardNode)
	app.Delete("/board/nodes/:id", api.DeleteBoardNode)
	app.Post("/board/edges", api.CreateBoardEdge)
	app.Put("/board/edges/:id", api.UpdateBoardEdge)
	app.Delete("/board/edges/:id", api.DeleteBoardEdge)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	input := func(kind, body string) collaboration.BoardNodeInput {
		return collaboration.BoardNodeInput{Kind: kind, Body: body, Color: "sand", BoardGeometry: collaboration.BoardGeometry{X: -200, Y: 100, Width: 320, Height: 240}}
	}
	node, e := s.SaveBoardNode(ctx, meta, 0, input("note", "shared note"))
	must(e)
	second, e := s.SaveBoardNode(ctx, meta, 0, input("card", "free card"))
	must(e)
	edgeInput := collaboration.BoardEdgeInput{SourceID: node.ID, TargetID: second.ID, SourceHandle: "right", TargetHandle: "left", Label: "related", Routing: "curve", Color: "blue", Arrow: true}
	edge, e := s.SaveBoardEdge(ctx, meta, 0, edgeInput)
	must(e)
	if _, e = s.SaveBoardEdge(ctx, meta, 0, edgeInput); e == nil || e.GetHTTPStatus() != 409 {
		t.Fatal("duplicate edge accepted")
	}
	edgeInput.TargetID = node.ID
	if _, e = s.SaveBoardEdge(ctx, meta, 0, edgeInput); e == nil || e.GetHTTPStatus() != 400 {
		t.Fatal("self edge accepted")
	}
	edgeInput.TargetID = 999999999
	if _, e = s.SaveBoardEdge(ctx, meta, 0, edgeInput); e == nil || e.GetHTTPStatus() != 409 {
		t.Fatal("dangling edge accepted")
	}
	edgeInput.TargetID = second.ID
	edgeInput.Version = edge.Version
	edgeInput.Routing = "step"
	edgeInput.Label = "new label"
	editedEdge, e := s.SaveBoardEdge(ctx, other, edge.ID, edgeInput)
	must(e)
	if _, e = s.SaveBoardEdge(ctx, meta, edge.ID, edgeInput); e == nil || e.GetHTTPStatus() != 409 {
		t.Fatal("stale edge update accepted")
	}
	if e = s.DeleteBoardEdge(ctx, meta, edge.ID, edge.Version); e == nil || e.GetHTTPStatus() != 409 {
		t.Fatal("stale edge delete accepted")
	}
	layouts := []collaboration.BoardLayout{{ID: node.ID, Version: node.Version, BoardGeometry: collaboration.BoardGeometry{X: 300, Y: -100, Width: 480, Height: 300, ZIndex: 4}}, {ID: second.ID, Version: second.Version, BoardGeometry: collaboration.BoardGeometry{X: 800, Y: 100, Width: 320, Height: 240, ZIndex: 2}}}
	count := queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_admin_audit_log WHERE resource LIKE 'collaboration.board_%'`)
	moved, e := s.MoveBoardNodes(ctx, other, layouts)
	must(e)
	if len(moved) != 2 || moved[0].Version != 2 || moved[0].X != 300 || moved[0].UpdatedByAccountID != 2 || queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_admin_audit_log WHERE resource LIKE 'collaboration.board_%'`) != count {
		t.Fatal("layout did not preserve version/audit contract")
	}
	layouts[0].Version = 2
	layouts[0].X = 500
	if _, e = s.MoveBoardNodes(ctx, meta, layouts); e == nil || e.GetHTTPStatus() != 409 {
		t.Fatal("stale grouped movement accepted")
	}
	if queryInt64(t, ctx, admin, `SELECT x FROM gfa_collaboration_board_node WHERE id=$1`, node.ID) != 300 {
		t.Fatal("grouped movement partially committed")
	}
	changed := input("note", "changed body")
	changed.Version = 2
	changedNode, e := s.SaveBoardNode(ctx, meta, node.ID, changed)
	must(e)
	if changedNode.X != 300 || changedNode.Version != 3 {
		t.Fatal("body update overwrote layout")
	}
	closeResponse(requestJSON(t, app, "PUT", fmt.Sprintf("/board/nodes/%d", node.ID), `{"kind":"note","body":"stale","width":320,"height":240,"version":1}`, nil, 409))
	closeResponse(requestJSON(t, app, "PUT", "/board/nodes/layout", fmt.Sprintf(`{"nodes":[{"id":%d,"version":3,"x":400,"width":320,"height":240,"body":"spoof"}]}`, node.ID), nil, 400))
	closeResponse(requestJSON(t, app, "DELETE", fmt.Sprintf("/board/nodes/%d", node.ID), `{"version":1}`, nil, 409))
	for _, kind := range []string{"text", "rectangle", "ellipse", "arrow"} {
		value := input(kind, "label")
		value.Rotation = 90
		annotation, err := s.SaveBoardNode(ctx, meta, 0, value)
		must(err)
		invalidEdge := edgeInput
		invalidEdge.TargetID = annotation.ID
		if _, err := s.SaveBoardEdge(ctx, meta, 0, invalidEdge); err == nil || err.GetHTTPStatus() != 409 {
			t.Fatal("annotation accepted as a connection endpoint")
		}
	}
	for _, ref := range []struct {
		kind string
		id   int64
	}{{"idea", ideaID}, {"game", gameID}, {"site", siteID}, {"game", gameID}} {
		value := input("card", "")
		value.ReferenceKind = ref.kind
		value.ReferenceID = ref.id
		_, e = s.SaveBoardNode(ctx, meta, 0, value)
		must(e)
	}
	invalid := input("card", "")
	invalid.ReferenceKind = "game"
	invalid.ReferenceID = 999999999
	if _, e = s.SaveBoardNode(ctx, meta, 0, invalid); e == nil || e.GetHTTPStatus() != 400 {
		t.Fatal("missing formal reference accepted")
	}
	trace.queries.Store(0)
	document, e := s.Board(ctx)
	must(e)
	if len(document.References) != 3 || trace.queries.Load() != 2 {
		t.Fatalf("unbounded reference reads or duplicates: refs=%d queries=%d", len(document.References), trace.queries.Load())
	}
	// Audited node/edge writes, including cascading deletion, must all roll back.
	_, err := admin.Exec(ctx, `ALTER TABLE gfa_admin_audit_log ADD CONSTRAINT canvas_test_audit_failure CHECK (action NOT LIKE 'collaboration.board_%') NOT VALID`)
	must(err)
	changed.Version = 3
	if _, e = s.SaveBoardNode(ctx, meta, node.ID, changed); e == nil {
		t.Fatal("node audit failure ignored")
	}
	if _, e = s.SaveBoardNode(ctx, meta, 0, input("note", "rollback")); e == nil {
		t.Fatal("node create audit failure ignored")
	}
	newEdge := edgeInput
	newEdge.SourceID, newEdge.TargetID = second.ID, node.ID
	if _, err := s.SaveBoardEdge(ctx, meta, 0, newEdge); err == nil {
		t.Fatal("edge create audit failure ignored")
	}
	if queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_collaboration_board_edge`) != 1 {
		t.Fatal("failed edge create persisted")
	}
	edgeInput.Version = editedEdge.Version
	if _, e = s.SaveBoardEdge(ctx, meta, edge.ID, edgeInput); e == nil {
		t.Fatal("edge audit failure ignored")
	}
	if e = s.DeleteBoardEdge(ctx, meta, edge.ID, editedEdge.Version); e == nil {
		t.Fatal("edge delete audit failure ignored")
	}
	if e = s.DeleteBoardNode(ctx, meta, node.ID, changedNode.Version); e == nil {
		t.Fatal("cascade audit failure ignored")
	}
	if queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_collaboration_board_edge WHERE id=$1`, edge.ID) != 1 || queryInt64(t, ctx, admin, `SELECT version FROM gfa_collaboration_board_node WHERE id=$1`, node.ID) != 3 {
		t.Fatal("audit failure changed canvas")
	}
	_, err = admin.Exec(ctx, `ALTER TABLE gfa_admin_audit_log DROP CONSTRAINT canvas_test_audit_failure`)
	must(err)
	standaloneEdge, e := s.SaveBoardEdge(ctx, meta, 0, newEdge)
	must(e)
	must(s.DeleteBoardEdge(ctx, meta, standaloneEdge.ID, standaloneEdge.Version))
	if queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_collaboration_board_edge WHERE id=$1`, standaloneEdge.ID) != 0 {
		t.Fatal("edge deletion did not persist")
	}
	must(s.DeleteBoardNode(ctx, meta, node.ID, changedNode.Version))
	if queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_collaboration_board_edge WHERE id=$1`, edge.ID) != 0 {
		t.Fatal("node deletion left dangling edge")
	}
	if queryInt64(t, ctx, admin, `SELECT count(*) FROM gfa_admin_audit_log WHERE action='collaboration.board_node.delete' AND before_data::jsonb->'edges'->0->>'label'='new label'`) != 1 {
		t.Fatal("cascade snapshot missing")
	}
	// Independent canvas nodes must remain editable without a board-wide revision.
	value := input("card", "independent edit")
	value.Version = 2
	_, e = s.SaveBoardNode(ctx, other, second.ID, value)
	must(e)
}
