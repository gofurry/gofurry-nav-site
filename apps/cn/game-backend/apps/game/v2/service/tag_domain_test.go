package service

import (
	"context"
	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
	"github.com/gofurry/gofurry-game-backend/common"
	"testing"
)

func TestRecommendationTagSemantics(t *testing.T) {
	for _, tc := range []struct {
		category, role string
		want           float64
	}{{"classification", "normal", 1.2}, {"species", "normal", 1.4}, {"platform", "normal", 0.4}, {"other", "normal", 1.3}, {"future", "normal", 1}, {"species", "primary", 2}, {"platform", "secondary", 1.5}} {
		for _, id := range []string{"1", "2014", "987654321"} {
			got := recommendationTagWeight(recommendationTag{ID: id, Code: "arbitrary", CategoryCode: tc.category, Role: tc.role})
			if got != tc.want {
				t.Fatalf("%+v id=%s weight=%f", tc, id, got)
			}
		}
	}
	if similarRecommendationAlgorithmVersion != "similar-v2.4.0-hybrid-cbf" {
		t.Fatal("algorithm version not bumped")
	}
}

type failedRebuildReader struct{ fakeDetailReader }

func (r *failedRebuildReader) RecomputeRecommendation(ctx context.Context, id int64, lang, region string, calculate func([]v2models.GameV2RecommendationFeature) ([]v2models.GfgGameV2Recommendation, common.GFError)) common.GFError {
	if id == 1 {
		return common.NewDaoError("test write failure")
	}
	return r.fakeDetailReader.RecomputeRecommendation(ctx, id, lang, region, calculate)
}
func TestRecommendationRebuildReportsFailures(t *testing.T) {
	reader := &failedRebuildReader{fakeDetailReader{features: []v2models.GameV2RecommendationFeature{{GameID: 1, Name: "Wolf"}, {GameID: 2, Name: "Wolf"}}}}
	result, err := NewReadModelServiceWithReader(reader).RebuildRecommendations(context.Background())
	if err == nil || result.Total != 2 || result.Rebuilt != 1 || result.Failed != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
