package details

import (
	"context"
	"errors"
	"strings"
	"testing"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/addons/assets"
	"github.com/gofurry/steam-go/web/storefront"

	"github.com/gofurry/gofurry-game-collector/collector/game/models"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/domain"
	steammapper "github.com/gofurry/gofurry-game-collector/collector/game/v2/mapper/steam"
	"github.com/gofurry/gofurry-game-collector/collector/game/v2/steamclient"
)

type recordingRepository struct {
	data  domain.DetailsCollection
	calls int
	err   error
}

func (repo *recordingRepository) SaveDetails(_ context.Context, data domain.DetailsCollection) error {
	repo.calls++
	repo.data = data
	return repo.err
}

func TestCollectGameRejectsMissingAdapter(t *testing.T) {
	t.Parallel()

	collector := NewCollector(nil, &recordingRepository{})
	result, err := collector.CollectGame(context.Background(), models.GameID{ID: 1, Appid: 550})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if result.Status != domain.StatusFailed {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

func TestRequestPlanKeepsCNBaseAndUSCanonicalAuthority(t *testing.T) {
	t.Parallel()
	collector := NewCollector(nil, &recordingRepository{})
	if len(collector.requests) != 3 || collector.requests[0].region != domain.RegionCN || !collector.requests[0].preferAsBase || collector.requests[0].canonical {
		t.Fatalf("unexpected CN base plan: %+v", collector.requests)
	}
	if collector.requests[1].region != domain.RegionUS || collector.requests[1].lang != domain.StoreLocaleEN || !collector.requests[1].canonical {
		t.Fatalf("unexpected US canonical plan: %+v", collector.requests)
	}
	if collector.requests[2].canonical {
		t.Fatalf("HK fallback must not be canonical: %+v", collector.requests[2])
	}
}

func TestCollectGameStoreBrowseSuccessAuthorizesScopes(t *testing.T) {
	repo := &recordingRepository{}
	collector := testCollector(repo, func(_ context.Context, _ uint32, plan requestPlan) ([]assets.URLItem, error) {
		return []assets.URLItem{{Kind: assets.KindLibraryCapsule, URL: "https://shared.akamai.steamstatic.com/" + string(plan.lang) + "/library_capsule.jpg"}}, nil
	})

	result, err := collector.CollectGame(context.Background(), models.GameID{ID: 1, Appid: 550})
	if err != nil || result.Status != domain.StatusSuccess {
		t.Fatalf("CollectGame result=%+v err=%v", result, err)
	}
	if repo.calls != 1 || !hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStorefront, "") || !hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStoreBrowse, "zh") || !hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStoreBrowse, "en") {
		t.Fatalf("unexpected saved replacement scopes: %+v", repo.data.AssetReplaceScopes)
	}
	if !hasVerticalCover(repo.data.Assets) {
		t.Fatalf("saved assets have no vertical cover: %+v", repo.data.Assets)
	}
}

func TestCollectGameStoreBrowseErrorIsPartialAndPreservesScope(t *testing.T) {
	repo := &recordingRepository{}
	collector := testCollector(repo, func(context.Context, uint32, requestPlan) ([]assets.URLItem, error) {
		return nil, &steam.APIError{Kind: steam.ErrorKindHTTPStatus, StatusCode: 429, Message: "rate limited"}
	})

	result, err := collector.CollectGame(context.Background(), models.GameID{ID: 1, Appid: 550})
	if err == nil || result.Status != domain.StatusPartial || result.Error == nil {
		t.Fatalf("CollectGame result=%+v err=%v", result, err)
	}
	if result.Error.Kind != "rate_limited" || result.UpstreamStatusCode != 429 || result.TrafficBucket != string(steamclient.BucketOfficialAPI) {
		t.Fatalf("unexpected task diagnostics: %+v", result)
	}
	if !strings.Contains(result.Error.Message, "appid=550") || !strings.Contains(result.Error.Message, "lang=zh") || !strings.Contains(result.Error.Message, "lang=en") {
		t.Fatalf("StoreBrowse error lost scope context: %q", result.Error.Message)
	}
	if repo.calls != 1 || !hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStorefront, "") || hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStoreBrowse, "zh") || hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStoreBrowse, "en") {
		t.Fatalf("failed StoreBrowse scope became authoritative: %+v", repo.data.AssetReplaceScopes)
	}
}

func TestCollectGameMissingVerticalCoverIsPartial(t *testing.T) {
	repo := &recordingRepository{}
	collector := testCollector(repo, func(context.Context, uint32, requestPlan) ([]assets.URLItem, error) {
		return []assets.URLItem{{Kind: assets.KindHeader, URL: "https://example.test/header.jpg"}}, nil
	})

	result, err := collector.CollectGame(context.Background(), models.GameID{ID: 1, Appid: 550})
	if err == nil || result.Status != domain.StatusPartial || result.Error == nil || !strings.Contains(result.Error.Message, "missing library_capsule/library_capsule_2x") {
		t.Fatalf("CollectGame result=%+v err=%v", result, err)
	}
	if hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStoreBrowse, "zh") || hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStoreBrowse, "en") {
		t.Fatalf("incomplete StoreBrowse response became authoritative: %+v", repo.data.AssetReplaceScopes)
	}
}

func TestCollectGameReplacesSuccessfulCNAndPreservesFailedEN(t *testing.T) {
	repo := &recordingRepository{}
	collector := testCollector(repo, func(_ context.Context, _ uint32, plan requestPlan) ([]assets.URLItem, error) {
		if plan.lang == domain.StoreLocaleEN {
			return nil, errors.New("temporary upstream failure")
		}
		return []assets.URLItem{{Kind: assets.KindLibraryCapsule2x, URL: "https://example.test/library_capsule_2x.jpg"}}, nil
	})

	result, err := collector.CollectGame(context.Background(), models.GameID{ID: 1, Appid: 550})
	if err == nil || result.Status != domain.StatusPartial {
		t.Fatalf("CollectGame result=%+v err=%v", result, err)
	}
	if !hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStoreBrowse, "zh") || hasScope(repo.data.AssetReplaceScopes, domain.AssetSourceStoreBrowse, "en") {
		t.Fatalf("partial StoreBrowse scopes = %+v", repo.data.AssetReplaceScopes)
	}
}

func TestFetchStoreBrowseAssetsUsesOfficialAPIBucket(t *testing.T) {
	wantErr := errors.New("stop before SDK call")
	var got steamclient.Bucket
	collector := &Collector{runSteam: func(_ context.Context, bucket steamclient.Bucket, _ func(context.Context, *steam.Client) error) error {
		got = bucket
		return wantErr
	}}

	_, err := collector.fetchStoreBrowseAssets(context.Background(), 550, requestPlan{region: domain.RegionCN, lang: domain.StoreLocaleZH, steamLang: "schinese"})
	if !errors.Is(err, wantErr) || got != steamclient.BucketOfficialAPI {
		t.Fatalf("bucket=%s err=%v, want bucket=%s err=%v", got, err, steamclient.BucketOfficialAPI, wantErr)
	}
}

func testCollector(repo *recordingRepository, storeBrowse func(context.Context, uint32, requestPlan) ([]assets.URLItem, error)) *Collector {
	collector := NewCollector(nil, repo)
	collector.runSteam = func(context.Context, steamclient.Bucket, func(context.Context, *steam.Client) error) error {
		return nil
	}
	collector.fetchAppDetailsFn = func(context.Context, uint32, requestPlan) (storefront.AppDetailsData, []byte, error) {
		return storefront.AppDetailsData{Type: "game", Name: "Test Game"}, []byte(`{"ok":true}`), nil
	}
	collector.fetchStoreBrowseFn = storeBrowse
	collector.mapper = steammapper.NewDetailsMapper()
	return collector
}

func hasScope(scopes []domain.AssetReplaceScope, source, language string) bool {
	for _, scope := range scopes {
		if scope.Source == source && scope.Language == language {
			return true
		}
	}
	return false
}
