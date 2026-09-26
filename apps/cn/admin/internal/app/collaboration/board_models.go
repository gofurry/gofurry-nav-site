package collaboration

import adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"

type BoardNode = adminsqlc.GfaCollaborationBoardNode
type BoardEdge = adminsqlc.GfaCollaborationBoardEdge
type BoardDocument struct {
	Nodes      []BoardNode      `json:"nodes"`
	Edges      []BoardEdge      `json:"edges"`
	References []BoardReference `json:"references"`
}
type BoardReference struct {
	Kind        string `json:"kind"`
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	ContentKind string `json:"content_kind"`
	Missing     bool   `json:"missing"`
}
type BoardGeometry struct {
	X      int32 `json:"x"`
	Y      int32 `json:"y"`
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
	ZIndex int32 `json:"z_index"`
}
type BoardNodeInput struct {
	BoardGeometry
	Kind          string `json:"kind"`
	Title         string `json:"title"`
	Body          string `json:"body"`
	Color         string `json:"color"`
	Rotation      int32  `json:"rotation"`
	ReferenceKind string `json:"reference_kind"`
	ReferenceID   int64  `json:"reference_id"`
	Version       int64  `json:"version"`
}
type BoardLayout struct {
	BoardGeometry
	ID      int64 `json:"id"`
	Version int64 `json:"version"`
}
type BoardEdgeInput struct {
	SourceID     int64  `json:"source_id"`
	TargetID     int64  `json:"target_id"`
	SourceHandle string `json:"source_handle"`
	TargetHandle string `json:"target_handle"`
	Routing      string `json:"routing"`
	Label        string `json:"label"`
	Color        string `json:"color"`
	Arrow        bool   `json:"arrow"`
	Version      int64  `json:"version"`
}
