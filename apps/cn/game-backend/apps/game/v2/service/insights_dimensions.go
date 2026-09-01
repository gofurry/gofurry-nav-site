package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"
)

var insightDimensionContracts = []v2models.InsightDimensionContract{
	{PublicKey: "primary_tag", InternalKey: "primary_tag_id", SliceMode: "partition"},
	{PublicKey: "tag", InternalKey: "tag_id", SliceMode: "overlapping"},
}

func (s *InsightsService) GetInsightsMetricBreakdown(ctx context.Context, publicKey, publicDimension string) (v2models.InsightDimensionBreakdown, error) {
	result := v2models.InsightDimensionBreakdown{
		Key: publicKey, Dimension: publicDimension, Items: []v2models.InsightDimensionSlice{},
	}
	metric, ok := resolveInsightMetric(publicKey)
	if !ok {
		return result, ErrInvalidInsightMetric
	}
	dimension, ok := resolveInsightDimension(publicDimension)
	if !ok {
		return result, ErrInvalidInsightDimension
	}
	result.SliceMode = dimension.SliceMode
	summary, err := s.store.GetInsightMetricSummary(ctx, metric)
	if err != nil || summary == nil {
		return result, err
	}
	asOf := insightFormatDate(summary.FactDate)
	result.AsOf = &asOf
	rows, err := s.store.ListInsightMetricBreakdown(ctx, metric, dimension, summary.FactDate)
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		result.Items = append(result.Items, insightPublicDimensionSlice(dimension, row))
	}
	return result, nil
}

func (s *InsightsService) GetInsightsMetricSliceTrend(ctx context.Context, publicKey, publicDimension, publicValue, requestedRange string) (v2models.InsightDimensionTrend, error) {
	result := v2models.InsightDimensionTrend{
		Key: publicKey, Dimension: publicDimension, RequestedRange: requestedRange,
		Points: []v2models.InsightDimensionTrendPoint{},
	}
	metric, ok := resolveInsightMetric(publicKey)
	if !ok {
		return result, ErrInvalidInsightMetric
	}
	dimension, ok := resolveInsightDimension(publicDimension)
	if !ok {
		return result, ErrInvalidInsightDimension
	}
	internalValue, normalizedValue, ok := normalizeInsightDimensionValue(dimension, publicValue)
	if !ok {
		return result, ErrInvalidInsightSlice
	}
	rangeDays, ok := parseInsightRange(requestedRange)
	if !ok {
		return result, ErrInvalidInsightRange
	}
	result.SliceMode = dimension.SliceMode
	result.Slice = insightPublicDimensionSliceRef(normalizedValue, nil, nil)
	summary, err := s.store.GetInsightMetricSummary(ctx, metric)
	if err != nil || summary == nil {
		return result, err
	}
	asOf := insightFormatDate(summary.FactDate)
	result.AsOf = &asOf
	availability, err := s.store.GetInsightMetricSliceAvailability(ctx, metric, dimension, internalValue)
	if err != nil {
		return result, err
	}
	result.AvailableFrom = insightDateStringPointer(availability.AvailableFrom)
	result.AvailableThrough = insightDateStringPointer(availability.AvailableThrough)
	result.Slice = insightPublicDimensionSliceRef(normalizedValue, availability.Label, availability.LabelEn)
	rows, err := s.store.ListInsightMetricSliceTrend(ctx, metric, dimension, internalValue, summary.FactDate, rangeDays)
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		known := row.PositiveCount + row.NegativeCount
		result.Points = append(result.Points, v2models.InsightDimensionTrendPoint{
			Date: insightFormatDate(row.FactDate), Population: row.Population, Eligible: row.Eligible, Known: known,
			MetricValue: insightRatio(row.PositiveCount, known), Coverage: insightRatio(known, row.Eligible),
		})
	}
	return result, nil
}

func resolveInsightDimension(publicKey string) (v2models.InsightDimensionContract, bool) {
	for _, contract := range insightDimensionContracts {
		if contract.PublicKey == publicKey {
			return contract, true
		}
	}
	return v2models.InsightDimensionContract{}, false
}

func insightPublicDimensionSlice(dimension v2models.InsightDimensionContract, row v2models.InsightDimensionRecord) v2models.InsightDimensionSlice {
	ref := insightPublicDimensionSliceRef(row.Value, row.Label, row.LabelEn)
	known := row.PositiveCount + row.NegativeCount
	return v2models.InsightDimensionSlice{
		Value: ref.Value, Label: ref.Label, LabelEn: ref.LabelEn,
		Population: row.Population, Eligible: row.Eligible, Known: known,
		MetricValue: insightRatio(row.PositiveCount, known), Coverage: insightRatio(known, row.Eligible),
	}
}

func insightPublicDimensionSliceRef(value string, label, labelEn *string) v2models.InsightDimensionSliceRef {
	if value == "unknown" {
		return v2models.InsightDimensionSliceRef{Value: value, Label: insightStringPointer("未知"), LabelEn: insightStringPointer("Unknown")}
	}
	if label == nil {
		label = insightStringPointer(fmt.Sprintf("标签 #%s", value))
	}
	if labelEn == nil {
		labelEn = insightStringPointer(fmt.Sprintf("Tag #%s", value))
	}
	return v2models.InsightDimensionSliceRef{Value: value, Label: label, LabelEn: labelEn}
}

func normalizeInsightDimensionValue(_ v2models.InsightDimensionContract, value string) (string, string, bool) {
	value = strings.TrimSpace(value)
	if value == "unknown" {
		return value, value, true
	}
	id, err := strconv.ParseInt(value, 10, 64)
	return value, value, err == nil && id > 0
}

func insightStringPointer(value string) *string { return &value }
