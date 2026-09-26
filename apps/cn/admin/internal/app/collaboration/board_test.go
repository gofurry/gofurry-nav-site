package collaboration

import (
	"strings"
	"testing"
)

func TestBoardNodeLimits(t *testing.T) {
	valid := BoardNodeInput{Kind: "note", Body: "note", BoardGeometry: BoardGeometry{X: -100, Y: -200, Width: 280, Height: 200}}
	if err := normalizeBoardNode(&valid); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*BoardNodeInput){
		func(in *BoardNodeInput) { in.Body = " " },
		func(in *BoardNodeInput) { in.X = -100001 },
		func(in *BoardNodeInput) { in.Height = 39 },
		func(in *BoardNodeInput) { in.Width = 1601 },
		func(in *BoardNodeInput) { in.ZIndex = 1000001 },
		func(in *BoardNodeInput) { in.Kind = "image" },
		func(in *BoardNodeInput) { in.Color = "red" },
		func(in *BoardNodeInput) { in.Rotation = 45 },
		func(in *BoardNodeInput) { in.Title = strings.Repeat("字", 201) },
		func(in *BoardNodeInput) { in.Body = strings.Repeat("字", 10001) },
		func(in *BoardNodeInput) { in.Body = "text\x00" },
		func(in *BoardNodeInput) { in.ReferenceKind = "game"; in.ReferenceID = 1 },
		func(in *BoardNodeInput) { in.Kind = "card"; in.ReferenceKind = "game"; in.ReferenceID = 0 },
		func(in *BoardNodeInput) { in.Kind = "card"; in.ReferenceKind = "invalid"; in.ReferenceID = 1 },
	} {
		in := valid
		change(&in)
		if normalizeBoardNode(&in) == nil {
			t.Fatalf("accepted %+v", in)
		}
	}
	for _, kind := range []string{"rectangle", "ellipse", "arrow"} {
		in := valid
		in.Kind = kind
		in.Body = ""
		if err := normalizeBoardNode(&in); err != nil {
			t.Fatalf("empty shape %s: %v", kind, err)
		}
	}
	in := valid
	in.Kind = "card"
	in.Body = ""
	in.ReferenceKind = "game"
	in.ReferenceID = 1
	if err := normalizeBoardNode(&in); err != nil {
		t.Fatal(err)
	}
}
