package controller

import (
	"reflect"
	"testing"

	navsqlc "github.com/gofurry/gofurry-admin/internal/db/nav/sqlc"
)

func TestCurationOrderRequiresCurrentCompleteMembership(t *testing.T) {
	rows := []navsqlc.ListGroupCurationRow{{ID: 1, SiteID: 11, Weight: 20}, {ID: 2, SiteID: 12, Weight: 10}}
	for _, tc := range []struct {
		name     string
		ids      []string
		revision string
		valid    bool
	}{
		{"reverse", []string{"12", "11"}, curationRevision(rows), true},
		{"duplicate", []string{"11", "11"}, curationRevision(rows), false},
		{"foreign", []string{"11", "13"}, curationRevision(rows), false},
		{"missing", []string{"11"}, curationRevision(rows), false},
		{"invalid", []string{"11", "oops"}, curationRevision(rows), false},
		{"stale", []string{"12", "11"}, "stale", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ids, err := validateCurationOrder(rows, curationOrder{Revision: tc.revision, SiteIDs: tc.ids})
			if (err == nil) != tc.valid {
				t.Fatalf("ids=%v err=%v", ids, err)
			}
			if tc.valid && !reflect.DeepEqual(ids, []int64{12, 11}) {
				t.Fatalf("order lost: %v", ids)
			}
		})
	}
	before := curationRevision(rows)
	rows[0].Weight++
	if before == curationRevision(rows) {
		t.Fatal("weight change must invalidate revision")
	}
	before = curationRevision(rows)
	rows[0].ID++
	if before == curationRevision(rows) {
		t.Fatal("bulk replacement must invalidate revision")
	}
}
