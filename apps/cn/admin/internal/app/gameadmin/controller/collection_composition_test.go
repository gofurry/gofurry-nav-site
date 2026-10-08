package controller

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/gofurry/gofurry-admin/internal/app/gameadmin/models"
)

func TestCompositionRequiredCanonicalDisjointArrays(t *testing.T) {
	for _, body := range []string{
		`{}`, `{"tag_ids":[],"manual_game_ids":[]}`, `{"tag_ids":null,"manual_game_ids":[],"excluded_game_ids":[]}`,
		`{"tag_ids":[0],"manual_game_ids":[],"excluded_game_ids":[]}`, `{"tag_ids":[],"manual_game_ids":[-1],"excluded_game_ids":[]}`,
		`{"tag_ids":[],"manual_game_ids":[2],"excluded_game_ids":[2]}`,
	} {
		var in models.ReplaceCollectionComposition
		if err := json.Unmarshal([]byte(body), &in); err != nil {
			t.Fatal(err)
		}
		if normalizeCollectionComposition(&in) == nil {
			t.Fatal("accepted invalid payload", body)
		}
	}
	in := models.ReplaceCollectionComposition{TagIDs: []int64{3, 1, 3}, ManualGameIDs: []int64{8, 4, 8}, ExcludedGameIDs: []int64{9, 2, 9}}
	if err := normalizeCollectionComposition(&in); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(in.TagIDs, []int64{1, 3}) || !reflect.DeepEqual(in.ManualGameIDs, []int64{4, 8}) || !reflect.DeepEqual(in.ExcludedGameIDs, []int64{2, 9}) {
		t.Fatal(in)
	}
}
func TestCompositionAuditBoundedWithoutDerivedMembers(t *testing.T) {
	c := models.CollectionComposition{Version: 7, Counts: models.CollectionCompositionCounts{AutoMatched: 10000, Effective: 10000}}
	for i := int64(1); i <= 300; i++ {
		c.ManualMembers = append(c.ManualMembers, models.CollectionMember{GameID: i})
		c.RuleTags = append(c.RuleTags, models.CollectionRuleTag{TagID: i})
		c.ExcludedMembers = append(c.ExcludedMembers, models.CollectionMember{GameID: i + 1000})
	}
	c.EffectiveMembers = []models.EffectiveCollectionMember{{CollectionMember: models.CollectionMember{Name: "derived-secret"}, Source: "automatic"}}
	a := compositionAudit(c)
	if len(a.ManualGameIDs) != 256 || len(a.TagIDs) != 256 || len(a.ExcludedGameIDs) != 256 || !a.IDsTruncated || a.Counts.Effective != 10000 {
		t.Fatal(a)
	}
	b, _ := json.Marshal(a)
	if strings.Contains(string(b), "derived-secret") || strings.Contains(string(b), "effective_members") {
		t.Fatal(string(b))
	}
	c.ManualMembers[299].GameID++
	if compositionAudit(c).ConfigurationSHA256 == a.ConfigurationSHA256 {
		t.Fatal("digest lost truncated IDs")
	}
}
func TestPublishedTextReadinessIgnoresPassiveMembership(t *testing.T) {
	c := models.CollectionContent{Name: "中文", NameEn: "English", Info: "简介", InfoEn: "Description"}
	if requireCollectionContent(c) != nil || requirePublishable(c, 1) == nil {
		t.Fatal("text readiness must be independent")
	}
	c.Info = ""
	if requireCollectionContent(c) == nil {
		t.Fatal("published text remains required")
	}
}
