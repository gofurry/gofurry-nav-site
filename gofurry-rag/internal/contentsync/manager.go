package contentsync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-rag/config"
	"github.com/gofurry/gofurry-rag/internal/db"
	"github.com/gofurry/gofurry-rag/internal/ingest"
)

const (
	SourceAll           = "all"
	SourceNavSites      = "nav_sites"
	SourceSiteChangelog = "site_changelog"
	SourceGameDetails   = "game_details"
	SourceGameNews      = "game_news"
	SourceGameCreators  = "game_creators"
	TriggerManual       = "manual"
	TriggerAuto         = "auto"
)

type Repository interface {
	UpsertSyncedDocument(ctx context.Context, params db.SyncDocumentParams) (db.SyncDocumentResult, error)
	CreateSyncRun(ctx context.Context, params db.CreateSyncRunParams) (db.SyncRun, error)
	CompleteSyncRun(ctx context.Context, id int64, params db.CompleteSyncRunParams) error
	LatestSyncRuns(ctx context.Context) (map[string]db.SyncRun, error)
	CountDocumentsBySourceType(ctx context.Context) (map[string]int64, error)
}

type NavClient interface {
	ListSites(ctx context.Context, locale string) ([]NavSite, error)
	ListGroups(ctx context.Context, locale string) ([]NavGroup, error)
	GetSiteDetail(ctx context.Context, id, locale string) (NavSiteDetail, error)
	GetSiteHTTP(ctx context.Context, domain string) (NavHTTPRecord, error)
	ListChangelogs(ctx context.Context) ([]ChangeLog, error)
	FetchMarkdown(ctx context.Context, rawURL string) (string, error)
}

type GameClient interface {
	ListGames(ctx context.Context, locale string) ([]GameSummary, error)
	GetGameInfo(ctx context.Context, id, locale string) (GameDetail, error)
	ListGameNews(ctx context.Context, locale string) ([]GameNews, error)
	ListCreators(ctx context.Context, locale string) ([]GameCreator, error)
}

type Manager struct {
	cfg        config.Config
	repo       Repository
	navClient  NavClient
	gameClient GameClient

	mu               sync.Mutex
	running          bool
	currentSource    string
	currentTrigger   string
	currentStartedAt *time.Time
	startOnce        sync.Once
}

type StatusResponse struct {
	Enabled          bool           `json:"enabled"`
	Running          bool           `json:"running"`
	CurrentSource    string         `json:"current_source,omitempty"`
	CurrentTrigger   string         `json:"current_trigger,omitempty"`
	CurrentStartedAt *time.Time     `json:"current_started_at,omitempty"`
	IntervalMinutes  int            `json:"interval_minutes"`
	Sources          []SourceStatus `json:"sources"`
}

type SourceStatus struct {
	Source               string      `json:"source"`
	Service              string      `json:"service"`
	AutoEnabled          bool        `json:"auto_enabled"`
	CurrentDocumentCount int64       `json:"current_document_count"`
	LastRun              *db.SyncRun `json:"last_run,omitempty"`
}

type syncError struct {
	status  int
	message string
}

func (e syncError) Error() string   { return e.message }
func (e syncError) HTTPStatus() int { return e.status }

func NewManager(cfg config.Config, repo Repository, navClient NavClient, gameClient GameClient) *Manager {
	if navClient == nil {
		navClient = NewHTTPNavClient(cfg.SyncNavBaseURL, time.Duration(cfg.SyncTimeoutSeconds)*time.Second)
	}
	if gameClient == nil {
		gameClient = NewHTTPGameClient(cfg.SyncGameBaseURL, time.Duration(cfg.SyncTimeoutSeconds)*time.Second)
	}
	return &Manager{cfg: cfg, repo: repo, navClient: navClient, gameClient: gameClient}
}

func (m *Manager) Start(ctx context.Context) {
	if m == nil || !m.cfg.SyncEnabled {
		return
	}
	m.startOnce.Do(func() {
		go m.scheduler(ctx)
	})
}

func (m *Manager) scheduler(ctx context.Context) {
	interval := time.Duration(maxInt(m.cfg.SyncIntervalMinutes, 1)) * time.Minute
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if err := m.Trigger(context.Background(), SourceAll, TriggerAuto); err != nil && !isSyncRunning(err) {
				slog.Warn("automatic sync trigger failed", "error", err)
			}
			timer.Reset(interval)
		}
	}
}

func (m *Manager) Trigger(ctx context.Context, source, trigger string) error {
	source, err := normalizeSource(source)
	if err != nil {
		return err
	}
	if trigger = strings.TrimSpace(trigger); trigger == "" {
		trigger = TriggerManual
	}
	startedAt, err := m.beginRun(source, trigger)
	if err != nil {
		return err
	}
	go m.execute(context.Background(), source, trigger, startedAt)
	return nil
}

func (m *Manager) RunNow(ctx context.Context, source, trigger string) error {
	source, err := normalizeSource(source)
	if err != nil {
		return err
	}
	if trigger = strings.TrimSpace(trigger); trigger == "" {
		trigger = TriggerManual
	}
	startedAt, err := m.beginRun(source, trigger)
	if err != nil {
		return err
	}
	return m.execute(ctx, source, trigger, startedAt)
}

func (m *Manager) Status(ctx context.Context) (StatusResponse, error) {
	runs, err := m.repo.LatestSyncRuns(ctx)
	if err != nil {
		return StatusResponse{}, err
	}
	documentCounts, err := m.repo.CountDocumentsBySourceType(ctx)
	if err != nil {
		return StatusResponse{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	resp := StatusResponse{
		Enabled:          m.cfg.SyncEnabled,
		Running:          m.running,
		CurrentSource:    m.currentSource,
		CurrentTrigger:   m.currentTrigger,
		CurrentStartedAt: m.currentStartedAt,
		IntervalMinutes:  maxInt(m.cfg.SyncIntervalMinutes, 1),
		Sources:          make([]SourceStatus, 0, len(sourceDefinitions())),
	}
	for _, item := range sourceDefinitions() {
		run, ok := runs[item.Source]
		var latest *db.SyncRun
		if ok {
			copied := run
			latest = &copied
		}
		resp.Sources = append(resp.Sources, SourceStatus{
			Source:               item.Source,
			Service:              item.Service,
			AutoEnabled:          m.cfg.SyncEnabled,
			CurrentDocumentCount: documentCounts[item.DocumentSource],
			LastRun:              latest,
		})
	}
	return resp, nil
}

func (m *Manager) execute(ctx context.Context, source, trigger string, startedAt time.Time) error {
	if ctx == nil {
		ctx = context.Background()
	}
	defer m.finishRun(startedAt)

	var errs []error
	for _, item := range expandSources(source) {
		m.setCurrentSource(startedAt, item)
		runCtx, cancel := context.WithTimeout(ctx, time.Duration(maxInt(m.cfg.SyncTimeoutSeconds, 1))*time.Second)
		err := m.runSource(runCtx, item, trigger)
		cancel()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (m *Manager) runSource(ctx context.Context, source, trigger string) error {
	runner, ok := sourceRunner(source)
	if !ok {
		return syncError{status: 400, message: "unsupported sync source"}
	}
	run, err := m.repo.CreateSyncRun(ctx, db.CreateSyncRunParams{Source: source, Trigger: trigger})
	if err != nil {
		return err
	}

	counts, runErr := runner(ctx, m)
	status := "success"
	message := ""
	if runErr != nil {
		message = runErr.Error()
		if counts.Added+counts.Updated+counts.Skipped > 0 || counts.Failed > 0 {
			status = "partial"
		} else {
			status = "failed"
		}
	} else if counts.Failed > 0 {
		status = "partial"
	}
	if err := m.repo.CompleteSyncRun(ctx, run.ID, db.CompleteSyncRunParams{
		Status:           status,
		SourceTotalCount: counts.Total,
		AddedCount:       counts.Added,
		UpdatedCount:     counts.Updated,
		SkippedCount:     counts.Skipped,
		FailedCount:      counts.Failed,
		Message:          message,
	}); err != nil {
		return err
	}
	return runErr
}

func (m *Manager) beginRun(source, trigger string) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.running {
		return time.Time{}, syncError{status: 409, message: "sync already running"}
	}
	now := time.Now()
	m.running = true
	m.currentSource = source
	m.currentTrigger = trigger
	m.currentStartedAt = &now
	return now, nil
}

func (m *Manager) finishRun(startedAt time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.currentStartedAt != nil && m.currentStartedAt.Equal(startedAt) {
		m.running = false
		m.currentSource = ""
		m.currentTrigger = ""
		m.currentStartedAt = nil
	}
}

func (m *Manager) setCurrentSource(startedAt time.Time, source string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.currentStartedAt != nil && m.currentStartedAt.Equal(startedAt) {
		m.currentSource = source
	}
}

type syncCounts struct {
	Total   int
	Added   int
	Updated int
	Skipped int
	Failed  int
}

type sourceDefinition struct {
	Source         string
	Service        string
	DocumentSource string
}

func sourceDefinitions() []sourceDefinition {
	return []sourceDefinition{
		{Source: SourceNavSites, Service: "gofurry-nav-backend", DocumentSource: "nav_site"},
		{Source: SourceSiteChangelog, Service: "gofurry-nav-backend", DocumentSource: "site_changelog"},
		{Source: SourceGameDetails, Service: "gofurry-game-backend", DocumentSource: "game_detail"},
		{Source: SourceGameNews, Service: "gofurry-game-backend", DocumentSource: "game_news"},
		{Source: SourceGameCreators, Service: "gofurry-game-backend", DocumentSource: "game_creator"},
	}
}

func sourceRunner(source string) (func(context.Context, *Manager) (syncCounts, error), bool) {
	switch source {
	case SourceNavSites:
		return runNavSitesSync, true
	case SourceSiteChangelog:
		return runChangeLogSync, true
	case SourceGameDetails:
		return runGameDetailsSync, true
	case SourceGameNews:
		return runGameNewsSync, true
	case SourceGameCreators:
		return runGameCreatorsSync, true
	default:
		return nil, false
	}
}

func expandSources(source string) []string {
	if source == SourceAll {
		items := sourceDefinitions()
		result := make([]string, 0, len(items))
		for _, item := range items {
			result = append(result, item.Source)
		}
		return result
	}
	return []string{source}
}

func normalizeSource(source string) (string, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		source = SourceAll
	}
	switch source {
	case SourceAll, SourceNavSites, SourceSiteChangelog, SourceGameDetails, SourceGameNews, SourceGameCreators:
		return source, nil
	default:
		return "", syncError{status: 400, message: "source must be one of nav_sites, site_changelog, game_details, game_news, game_creators, all"}
	}
}

func isSyncRunning(err error) bool {
	var se syncError
	return errors.As(err, &se) && se.status == 409
}

func runNavSitesSync(ctx context.Context, m *Manager) (syncCounts, error) {
	locales := []string{"zh-CN", "en-US"}
	var counts syncCounts
	var errs []error
	for _, locale := range locales {
		groups, err := m.navClient.ListGroups(ctx, locale)
		if err != nil {
			errs = append(errs, fmt.Errorf("load %s groups: %w", locale, err))
			continue
		}
		sites, err := m.navClient.ListSites(ctx, locale)
		if err != nil {
			errs = append(errs, fmt.Errorf("load %s sites: %w", locale, err))
			continue
		}
		siteGroups := buildSiteGroups(groups)
		for _, site := range sites {
			counts.Total++
			detail, err := m.navClient.GetSiteDetail(ctx, site.ID, locale)
			if err != nil {
				counts.Failed++
				errs = append(errs, fmt.Errorf("load site detail %s/%s: %w", site.ID, locale, err))
				continue
			}
			httpRecord, httpErr := m.navClient.GetSiteHTTP(ctx, site.Domain)
			if httpErr != nil {
				slog.Warn("load site http record failed", "site_id", site.ID, "locale", locale, "domain", site.Domain, "error", httpErr)
			}
			metadata, content, title, targetURL, checksum, err := buildNavSitePayload(site, detail, siteGroups[site.ID], httpRecord, locale)
			if err != nil {
				counts.Failed++
				errs = append(errs, fmt.Errorf("build site payload %s/%s: %w", site.ID, locale, err))
				continue
			}
			result, err := m.repo.UpsertSyncedDocument(ctx, db.SyncDocumentParams{
				Title:      title,
				Content:    content,
				SourceType: "nav_site",
				SourceID:   fmt.Sprintf("nav-site:%s:%s", site.ID, locale),
				URL:        targetURL,
				Checksum:   checksum,
				Metadata:   metadata,
			})
			if err != nil {
				counts.Failed++
				errs = append(errs, fmt.Errorf("upsert site document %s/%s: %w", site.ID, locale, err))
				continue
			}
			applySyncAction(&counts, result.Action)
		}
	}
	return counts, errors.Join(errs...)
}

func runChangeLogSync(ctx context.Context, m *Manager) (syncCounts, error) {
	var counts syncCounts
	list, err := m.navClient.ListChangelogs(ctx)
	if err != nil {
		return counts, err
	}
	var errs []error
	for _, item := range list {
		counts.Total++
		text, err := m.navClient.FetchMarkdown(ctx, item.URL)
		if err != nil {
			counts.Failed++
			errs = append(errs, fmt.Errorf("load changelog %s: %w", item.URL, err))
			continue
		}
		metadata, content, checksum, err := buildChangeLogPayload(item, text)
		if err != nil {
			counts.Failed++
			errs = append(errs, fmt.Errorf("build changelog payload %s: %w", item.URL, err))
			continue
		}
		result, err := m.repo.UpsertSyncedDocument(ctx, db.SyncDocumentParams{
			Title:      strings.TrimSpace(item.Title),
			Content:    content,
			SourceType: "site_changelog",
			SourceID:   "site-changelog:" + strings.TrimSpace(item.URL),
			URL:        strings.TrimSpace(item.URL),
			Checksum:   checksum,
			Metadata:   metadata,
		})
		if err != nil {
			counts.Failed++
			errs = append(errs, fmt.Errorf("upsert changelog %s: %w", item.URL, err))
			continue
		}
		applySyncAction(&counts, result.Action)
	}
	return counts, errors.Join(errs...)
}

func applySyncAction(counts *syncCounts, action string) {
	switch action {
	case "created":
		counts.Added++
	case "updated":
		counts.Updated++
	default:
		counts.Skipped++
	}
}

type navGroupRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func buildSiteGroups(groups []NavGroup) map[string][]navGroupRef {
	result := make(map[string][]navGroupRef)
	for _, group := range groups {
		ref := navGroupRef{ID: strings.TrimSpace(group.ID), Name: strings.TrimSpace(group.Name)}
		for _, siteID := range group.Sites {
			siteID = strings.TrimSpace(siteID)
			if siteID == "" {
				continue
			}
			result[siteID] = append(result[siteID], ref)
		}
	}
	return result
}

func buildNavSitePayload(site NavSite, detail NavSiteDetail, groups []navGroupRef, httpRecord NavHTTPRecord, locale string) (json.RawMessage, string, string, string, string, error) {
	targetURL := strings.TrimSpace(httpRecord.URL)
	if targetURL == "" && strings.TrimSpace(site.Domain) != "" {
		targetURL = "https://" + strings.TrimSpace(site.Domain)
	}
	groupIDs := make([]string, 0, len(groups))
	groupNames := make([]string, 0, len(groups))
	for _, group := range groups {
		if group.ID != "" {
			groupIDs = append(groupIDs, group.ID)
		}
		if group.Name != "" {
			groupNames = append(groupNames, group.Name)
		}
	}
	metadataMap := map[string]any{
		"category":    "nav",
		"language":    locale,
		"site_id":     strings.TrimSpace(site.ID),
		"domain":      strings.TrimSpace(site.Domain),
		"group_ids":   groupIDs,
		"group_names": groupNames,
		"country":     firstNonEmpty(detail.Country, site.Country),
		"nsfw":        firstNonEmpty(detail.NSFW, site.NSFW),
		"welfare":     firstNonEmpty(detail.Welfare, site.Welfare),
	}
	metadata, err := json.Marshal(metadataMap)
	if err != nil {
		return nil, "", "", "", "", err
	}
	title := strings.TrimSpace(site.Name)
	if title == "" {
		title = strings.TrimSpace(detail.Name)
	}
	content := renderNavSiteContent(site, detail, groups, httpRecord, targetURL, locale)
	checksum := syncChecksum(title, targetURL, content, metadata)
	return metadata, content, title, targetURL, checksum, nil
}

func buildChangeLogPayload(item ChangeLog, markdown string) (json.RawMessage, string, string, error) {
	metadata, err := json.Marshal(map[string]any{
		"category":     "changelog",
		"language":     "multi",
		"path":         "/updates",
		"published_at": strings.TrimSpace(item.CreateTime),
		"updated_at":   strings.TrimSpace(item.UpdateTime),
	})
	if err != nil {
		return nil, "", "", err
	}
	content := renderChangeLogContent(item, markdown)
	return metadata, content, syncChecksum(strings.TrimSpace(item.Title), strings.TrimSpace(item.URL), content, metadata), nil
}

func renderNavSiteContent(site NavSite, detail NavSiteDetail, groups []navGroupRef, httpRecord NavHTTPRecord, targetURL, locale string) string {
	groupNames := make([]string, 0, len(groups))
	for _, group := range groups {
		if group.Name != "" {
			groupNames = append(groupNames, group.Name)
		}
	}
	name := firstNonEmpty(site.Name, detail.Name)
	info := firstNonEmpty(site.Info, detail.Info)
	country := firstNonEmpty(detail.Country, site.Country)
	nsfw := firstNonEmpty(detail.NSFW, site.NSFW)
	welfare := firstNonEmpty(detail.Welfare, site.Welfare)
	pageTitle := strings.TrimSpace(httpRecord.Title)
	pageDescription := strings.TrimSpace(httpRecord.Meta.Description)

	var builder strings.Builder
	if locale == "en-US" {
		writeSection(&builder, "Name", name)
		writeSection(&builder, "Domain", site.Domain)
		writeSection(&builder, "Description", info)
		writeSection(&builder, "Groups", strings.Join(groupNames, ", "))
		writeSection(&builder, "Country", country)
		writeSection(&builder, "NSFW", nsfw)
		writeSection(&builder, "Welfare", welfare)
		writeSection(&builder, "Page Title", pageTitle)
		writeSection(&builder, "Page Description", pageDescription)
		writeSection(&builder, "URL", targetURL)
		return strings.TrimSpace(builder.String())
	}
	writeSection(&builder, "站点名", name)
	writeSection(&builder, "域名", site.Domain)
	writeSection(&builder, "简介", info)
	writeSection(&builder, "所属分组", strings.Join(groupNames, "、"))
	writeSection(&builder, "国家", country)
	writeSection(&builder, "NSFW", nsfw)
	writeSection(&builder, "福利属性", welfare)
	writeSection(&builder, "页面标题", pageTitle)
	writeSection(&builder, "页面描述", pageDescription)
	writeSection(&builder, "访问链接", targetURL)
	return strings.TrimSpace(builder.String())
}

func renderChangeLogContent(item ChangeLog, markdown string) string {
	var builder strings.Builder
	writeSection(&builder, "标题", item.Title)
	writeSection(&builder, "创建时间", item.CreateTime)
	writeSection(&builder, "更新时间", item.UpdateTime)
	writeSection(&builder, "原始链接", item.URL)
	body := strings.TrimSpace(markdown)
	if body != "" {
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(body)
	}
	return strings.TrimSpace(builder.String())
}

func writeSection(builder *strings.Builder, label, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if builder.Len() > 0 {
		builder.WriteString("\n\n")
	}
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(value)
}

func syncChecksum(title, targetURL, content string, metadata json.RawMessage) string {
	payload := strings.Join([]string{
		strings.TrimSpace(title),
		strings.TrimSpace(targetURL),
		strings.TrimSpace(content),
		strings.TrimSpace(string(metadata)),
	}, "\n---\n")
	return ingest.Checksum(payload)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func maxInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
