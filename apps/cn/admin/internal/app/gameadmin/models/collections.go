package models

import "time"

// GameCollection is the Admin read model; membership has no editorial order.
type GameCollection struct {
	ID             int64      `json:"id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	NameEn         string     `json:"name_en"`
	Info           string     `json:"info"`
	InfoEn         string     `json:"info_en"`
	Status         string     `json:"status"`
	Version        int64      `json:"version"`
	PublishedAt    *time.Time `json:"published_at"`
	ArchivedAt     *time.Time `json:"archived_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	MemberCount    int64      `json:"member_count"`
	SFWMemberCount int64      `json:"sfw_member_count"`
	HomeSlot       *int16     `json:"home_slot"`
}

type CollectionContent struct {
	Name   string `json:"name"`
	NameEn string `json:"name_en"`
	Info   string `json:"info"`
	InfoEn string `json:"info_en"`
}
type CreateCollection struct {
	Code string `json:"code"`
	CollectionContent
}
type UpdateCollection struct {
	Version int64 `json:"version"`
	CollectionContent
}
type CollectionVersion struct {
	Version int64 `json:"version"`
}
type ReplaceCollectionMembers struct {
	Version int64   `json:"version"`
	GameIDs []int64 `json:"game_ids"`
}
type CollectionMember struct {
	GameID int64  `json:"game_id"`
	Name   string `json:"name"`
	NameEn string `json:"name_en"`
	AppID  int64  `json:"appid"`
	Adult  bool   `json:"adult"`
}
type CollectionMembers struct {
	CollectionID int64              `json:"collection_id"`
	Version      int64              `json:"version"`
	Members      []CollectionMember `json:"members"`
}
type CollectionHomeSlot struct {
	Slot       int16           `json:"slot"`
	Collection *GameCollection `json:"collection"`
}
type CollectionHome struct {
	Revision string               `json:"revision"`
	Slots    []CollectionHomeSlot `json:"slots"`
}
type CollectionPlacement struct {
	Slot         int16  `json:"slot"`
	CollectionID *int64 `json:"collection_id"`
}
type ReplaceCollectionHome struct {
	Revision string                `json:"revision"`
	Slots    []CollectionPlacement `json:"slots"`
}

// Composition is separate from the legacy manual-only members API.
type CollectionRuleTag struct {
	TagID  int64  `json:"tag_id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	NameEn string `json:"name_en"`
	Active bool   `json:"active"`
}
type EffectiveCollectionMember struct {
	CollectionMember
	Source string `json:"source"`
}
type CollectionCompositionCounts struct {
	AutoMatched  int64 `json:"auto_matched"`
	ManualPinned int64 `json:"manual_pinned"`
	Excluded     int64 `json:"excluded"`
	Effective    int64 `json:"effective"`
	SFWVisible   int64 `json:"sfw_visible"`
}
type CollectionComposition struct {
	CollectionID     int64                       `json:"collection_id"`
	Version          int64                       `json:"version"`
	HomeSlot         *int16                      `json:"home_slot"`
	RuleTags         []CollectionRuleTag         `json:"rule_tags"`
	ManualMembers    []CollectionMember          `json:"manual_members"`
	ExcludedMembers  []CollectionMember          `json:"excluded_members"`
	EffectiveMembers []EffectiveCollectionMember `json:"effective_members"`
	Counts           CollectionCompositionCounts `json:"counts"`
}
type ReplaceCollectionComposition struct {
	Version         int64   `json:"version"`
	TagIDs          []int64 `json:"tag_ids"`
	ManualGameIDs   []int64 `json:"manual_game_ids"`
	ExcludedGameIDs []int64 `json:"excluded_game_ids"`
}
