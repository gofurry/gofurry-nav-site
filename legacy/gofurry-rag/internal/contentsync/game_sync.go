package contentsync

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gofurry/gofurry-rag/internal/db"
)

func runGameDetailsSync(ctx context.Context, m *Manager) (syncCounts, error) {
	locales := []string{"zh", "en"}
	var counts syncCounts
	var errs []error
	for _, locale := range locales {
		games, err := m.gameClient.ListGames(ctx, locale)
		if err != nil {
			errs = append(errs, fmt.Errorf("load %s game list: %w", locale, err))
			continue
		}
		for _, game := range games {
			counts.Total++
			detail, err := m.gameClient.GetGameInfo(ctx, game.ID, locale)
			if err != nil {
				counts.Failed++
				errs = append(errs, fmt.Errorf("load game detail %s/%s: %w", game.ID, locale, err))
				continue
			}
			metadata, content, title, targetURL, checksum, err := buildGameDetailPayload(game, detail, locale)
			if err != nil {
				counts.Failed++
				errs = append(errs, fmt.Errorf("build game detail payload %s/%s: %w", game.ID, locale, err))
				continue
			}
			result, err := m.repo.UpsertSyncedDocument(ctx, db.SyncDocumentParams{
				Title:      title,
				Content:    content,
				SourceType: "game_detail",
				SourceID:   fmt.Sprintf("game:%s:%s", strings.TrimSpace(game.ID), locale),
				URL:        targetURL,
				Checksum:   checksum,
				Metadata:   metadata,
			})
			if err != nil {
				counts.Failed++
				errs = append(errs, fmt.Errorf("upsert game detail %s/%s: %w", game.ID, locale, err))
				continue
			}
			applySyncAction(&counts, result.Action)
		}
	}
	return counts, joinSyncErrors(errs)
}

func runGameNewsSync(ctx context.Context, m *Manager) (syncCounts, error) {
	locales := []string{"zh", "en"}
	var counts syncCounts
	var errs []error
	for _, locale := range locales {
		items, err := m.gameClient.ListGameNews(ctx, locale)
		if err != nil {
			errs = append(errs, fmt.Errorf("load %s game news: %w", locale, err))
			continue
		}
		for _, item := range items {
			counts.Total++
			metadata, content, title, targetURL, sourceID, checksum, err := buildGameNewsPayload(item, locale)
			if err != nil {
				counts.Failed++
				errs = append(errs, fmt.Errorf("build game news payload %s/%s: %w", item.Headline, locale, err))
				continue
			}
			result, err := m.repo.UpsertSyncedDocument(ctx, db.SyncDocumentParams{
				Title:      title,
				Content:    content,
				SourceType: "game_news",
				SourceID:   sourceID,
				URL:        targetURL,
				Checksum:   checksum,
				Metadata:   metadata,
			})
			if err != nil {
				counts.Failed++
				errs = append(errs, fmt.Errorf("upsert game news %s/%s: %w", item.Headline, locale, err))
				continue
			}
			applySyncAction(&counts, result.Action)
		}
	}
	return counts, joinSyncErrors(errs)
}

func buildGameDetailPayload(summary GameSummary, detail GameDetail, locale string) (json.RawMessage, string, string, string, string, error) {
	groupNames := kvValues(detail.Groups)
	tagNames := tagNames(detail.Tags)
	developers := fallbackStrings(detail.Developers, summary.Developers)
	publishers := fallbackStrings(detail.Publishers, summary.Publishers)
	title := firstNonEmpty(detail.Name, summary.Name)
	targetURL := firstNonEmpty(detail.Website, firstKVValue(detail.Links), firstKVValue(detail.Resources))
	metadata, err := json.Marshal(map[string]any{
		"category":    "game",
		"language":    locale,
		"game_id":     strings.TrimSpace(summary.ID),
		"platform":    strings.TrimSpace(detail.Platform),
		"group_names": groupNames,
		"tag_names":   tagNames,
		"developers":  developers,
		"publishers":  publishers,
		"website":     strings.TrimSpace(detail.Website),
	})
	if err != nil {
		return nil, "", "", "", "", err
	}
	content := renderGameDetailContent(summary, detail, locale)
	checksum := syncChecksum(title, targetURL, content, metadata)
	return metadata, content, title, targetURL, checksum, nil
}

func buildGameNewsPayload(item GameNews, locale string) (json.RawMessage, string, string, string, string, string, error) {
	title := strings.TrimSpace(item.Headline)
	if title == "" {
		title = strings.TrimSpace(item.Name)
	}
	targetURL := strings.TrimSpace(item.URL)
	metadata, err := json.Marshal(map[string]any{
		"category":     "game_news",
		"language":     locale,
		"game_name":    strings.TrimSpace(item.Name),
		"published_at": strings.TrimSpace(item.PostTime),
		"author":       strings.TrimSpace(item.Author),
		"url":          targetURL,
	})
	if err != nil {
		return nil, "", "", "", "", "", err
	}
	content := renderGameNewsContent(item, locale)
	identity := targetURL
	if identity == "" {
		identity = strings.Join([]string{strings.TrimSpace(item.Headline), strings.TrimSpace(item.PostTime), strings.TrimSpace(item.Name)}, "\n")
	}
	sourceID := fmt.Sprintf("game-news:%s:%s", locale, md5Hex(identity))
	checksum := syncChecksum(title, targetURL, content, metadata)
	return metadata, content, title, targetURL, sourceID, checksum, nil
}

func renderGameDetailContent(summary GameSummary, detail GameDetail, locale string) string {
	developers := strings.Join(fallbackStrings(detail.Developers, summary.Developers), joinToken(locale))
	publishers := strings.Join(fallbackStrings(detail.Publishers, summary.Publishers), joinToken(locale))
	groupNames := strings.Join(kvValues(detail.Groups), joinToken(locale))
	tagNames := strings.Join(tagNames(detail.Tags), joinToken(locale))
	resources := joinMultiline(detail.Resources)
	links := joinMultiline(detail.Links)
	reqs := joinRequirements(detail.PcRequirements, locale)
	name := firstNonEmpty(detail.Name, summary.Name)
	info := firstNonEmpty(detail.Info, summary.Info)

	var builder strings.Builder
	if locale == "en" {
		writeSection(&builder, "Name", name)
		writeSection(&builder, "Summary", info)
		writeSection(&builder, "Detailed Description", detail.DetailedDescription)
		writeSection(&builder, "About the Game", detail.AboutTheGame)
		writeSection(&builder, "Platform", detail.Platform)
		writeSection(&builder, "Supported Languages", detail.SupportedLanguages)
		writeSection(&builder, "Release Date", firstNonEmpty(detail.ReleaseDate, summary.ReleaseDate))
		writeSection(&builder, "Developers", developers)
		writeSection(&builder, "Publishers", publishers)
		writeSection(&builder, "Groups", groupNames)
		writeSection(&builder, "Tags", tagNames)
		writeSection(&builder, "Website", detail.Website)
		writeSection(&builder, "Resources", resources)
		writeSection(&builder, "Links", links)
		writeSection(&builder, "PC Requirements", reqs)
		return strings.TrimSpace(builder.String())
	}
	writeSection(&builder, "游戏名", name)
	writeSection(&builder, "简介", info)
	writeSection(&builder, "详细介绍", detail.DetailedDescription)
	writeSection(&builder, "关于游戏", detail.AboutTheGame)
	writeSection(&builder, "平台", detail.Platform)
	writeSection(&builder, "支持语言", detail.SupportedLanguages)
	writeSection(&builder, "发售日期", firstNonEmpty(detail.ReleaseDate, summary.ReleaseDate))
	writeSection(&builder, "开发者", developers)
	writeSection(&builder, "发行商", publishers)
	writeSection(&builder, "所属分组", groupNames)
	writeSection(&builder, "标签", tagNames)
	writeSection(&builder, "官网", detail.Website)
	writeSection(&builder, "资源链接", resources)
	writeSection(&builder, "相关链接", links)
	writeSection(&builder, "PC 配置要求", reqs)
	return strings.TrimSpace(builder.String())
}

func renderGameNewsContent(item GameNews, locale string) string {
	var builder strings.Builder
	if locale == "en" {
		writeSection(&builder, "Game", item.Name)
		writeSection(&builder, "Headline", item.Headline)
		writeSection(&builder, "Published At", item.PostTime)
		writeSection(&builder, "Author", item.Author)
		writeSection(&builder, "Original URL", item.URL)
	} else {
		writeSection(&builder, "游戏", item.Name)
		writeSection(&builder, "标题", item.Headline)
		writeSection(&builder, "发布时间", item.PostTime)
		writeSection(&builder, "作者", item.Author)
		writeSection(&builder, "原始链接", item.URL)
	}
	body := strings.TrimSpace(item.Content)
	if body != "" {
		if builder.Len() > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(body)
	}
	return strings.TrimSpace(builder.String())
}

func kvValues(items []GameKV) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		value := firstNonEmpty(item.Value, item.Key)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func tagNames(items []GameTag) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if name := strings.TrimSpace(item.Name); name != "" {
			result = append(result, name)
		}
	}
	return result
}

func joinMultiline(items []GameKV) string {
	parts := make([]string, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Key)
		value := strings.TrimSpace(item.Value)
		switch {
		case label != "" && value != "":
			parts = append(parts, label+": "+value)
		case value != "":
			parts = append(parts, value)
		case label != "":
			parts = append(parts, label)
		}
	}
	return strings.Join(parts, "\n")
}

func joinRequirements(req GamePCRequirements, locale string) string {
	minimum := strings.TrimSpace(req.Minimum)
	recommended := strings.TrimSpace(req.Recommended)
	if minimum == "" && recommended == "" {
		return ""
	}
	var parts []string
	if locale == "en" {
		if minimum != "" {
			parts = append(parts, "Minimum: "+minimum)
		}
		if recommended != "" {
			parts = append(parts, "Recommended: "+recommended)
		}
	} else {
		if minimum != "" {
			parts = append(parts, "最低配置: "+minimum)
		}
		if recommended != "" {
			parts = append(parts, "推荐配置: "+recommended)
		}
	}
	return strings.Join(parts, "\n")
}

func firstKVValue(items []GameKV) string {
	for _, item := range items {
		if value := strings.TrimSpace(item.Value); value != "" {
			return value
		}
	}
	return ""
}

func fallbackStrings(primary, secondary []string) []string {
	values := primary
	if len(values) == 0 {
		values = secondary
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func joinToken(locale string) string {
	if locale == "en" {
		return ", "
	}
	return "、"
}

func md5Hex(text string) string {
	sum := md5.Sum([]byte(strings.TrimSpace(text)))
	return hex.EncodeToString(sum[:])
}

func joinSyncErrors(errs []error) error {
	return errors.Join(errs...)
}
