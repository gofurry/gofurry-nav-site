package collaboration

import adminsqlc "github.com/gofurry/gofurry-admin/internal/db/admin/sqlc"

type Idea = adminsqlc.GetContentIdeaRow
type BoardNote = adminsqlc.GfaCollaborationBoardNote
type Summary = adminsqlc.CountContentIdeaSummaryRow

type Input struct {
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	Note     string `json:"note"`
	Priority string `json:"priority"`
	Version  int64  `json:"version"`
}
type Transition struct {
	Version    int64  `json:"version"`
	Kind       string `json:"kind,omitempty"`
	ResourceID int64  `json:"resource_id,omitempty"`
}
type BatchInput struct {
	Items     []Input `json:"items"`
	SkipKnown *bool   `json:"skip_known,omitempty"`
}
type Match struct {
	ID    int64  `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
}
type PreviewRow struct {
	Input
	Index            int      `json:"index"`
	SourceKey        *string  `json:"source_key"`
	DisplayHint      string   `json:"display_hint"`
	Valid            bool     `json:"valid"`
	Errors           []string `json:"errors"`
	Warnings         []string `json:"warnings"`
	IdeaMatch        *Match   `json:"idea_match"`
	ResourceMatch    *Match   `json:"resource_match"`
	BatchDuplicateOf *int     `json:"batch_duplicate_of"`
}
type BatchResult struct {
	InsertedCount int64 `json:"inserted_count"`
	SkippedCount  int   `json:"skipped_count"`
}
type Filters struct {
	Kind, Status, Priority, Researcher, Keyword string
	Page, PageSize                              int
}
type IdeaPage struct {
	Total int64                           `json:"total"`
	List  []adminsqlc.ListContentIdeasRow `json:"list"`
}
type BoardInput struct {
	Body    string `json:"body"`
	X       int32  `json:"x"`
	Y       int32  `json:"y"`
	Width   int32  `json:"width"`
	Height  int32  `json:"height"`
	ZIndex  int32  `json:"z_index"`
	Version int64  `json:"version"`
}
