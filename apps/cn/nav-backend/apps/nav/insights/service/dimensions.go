package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/insights/models"
)

var dimensionContracts = []models.DimensionContract{
	{PublicKey: "country", InternalKey: "site_country", SliceMode: "partition"},
	{PublicKey: "group", InternalKey: "group_id", SliceMode: "overlapping"},
	{PublicKey: "nsfw", InternalKey: "nsfw", SliceMode: "partition"},
	{PublicKey: "public_interest", InternalKey: "welfare", SliceMode: "partition"},
}

func (s *InsightsService) GetMetricBreakdown(ctx context.Context, publicKey, publicDimension string) (models.DimensionBreakdown, error) {
	result := models.DimensionBreakdown{
		Key: publicKey, Dimension: publicDimension, Items: []models.DimensionSlice{},
	}
	metric, ok := resolveMetric(publicKey)
	if !ok {
		return result, ErrInvalidMetricKey
	}
	dimension, ok := resolveDimension(publicDimension)
	if !ok {
		return result, ErrInvalidDimension
	}
	result.SliceMode = dimension.SliceMode
	summary, err := s.store.GetMetricSummary(ctx, metric)
	if err != nil || summary == nil {
		return result, err
	}
	asOf := formatDate(summary.FactDate)
	result.AsOf = &asOf
	rows, err := s.store.ListMetricBreakdown(ctx, metric, dimension, summary.FactDate)
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		result.Items = append(result.Items, publicDimensionSlice(dimension, row))
	}
	return result, nil
}

func (s *InsightsService) GetMetricSliceTrend(ctx context.Context, publicKey, publicDimension, publicValue, requestedRange string) (models.DimensionTrend, error) {
	result := models.DimensionTrend{
		Key: publicKey, Dimension: publicDimension, RequestedRange: requestedRange,
		Points: []models.DimensionTrendPoint{},
	}
	metric, ok := resolveMetric(publicKey)
	if !ok {
		return result, ErrInvalidMetricKey
	}
	dimension, ok := resolveDimension(publicDimension)
	if !ok {
		return result, ErrInvalidDimension
	}
	internalValue, normalizedValue, ok := normalizeDimensionValue(dimension, publicValue)
	if !ok {
		return result, ErrInvalidSlice
	}
	rangeDays, ok := parseRange(requestedRange)
	if !ok {
		return result, ErrInvalidRange
	}
	result.SliceMode = dimension.SliceMode
	result.Slice = publicDimensionSliceRef(dimension, normalizedValue, nil, nil)
	summary, err := s.store.GetMetricSummary(ctx, metric)
	if err != nil || summary == nil {
		return result, err
	}
	asOf := formatDate(summary.FactDate)
	result.AsOf = &asOf
	availability, err := s.store.GetMetricSliceAvailability(ctx, metric, dimension, internalValue)
	if err != nil {
		return result, err
	}
	result.AvailableFrom = dateStringPointer(availability.AvailableFrom)
	result.AvailableThrough = dateStringPointer(availability.AvailableThrough)
	result.Slice = publicDimensionSliceRef(dimension, normalizedValue, availability.Label, availability.LabelEn)
	rows, err := s.store.ListMetricSliceTrend(ctx, metric, dimension, internalValue, summary.FactDate, rangeDays)
	if err != nil {
		return result, err
	}
	for _, row := range rows {
		known := row.PositiveCount + row.NegativeCount
		result.Points = append(result.Points, models.DimensionTrendPoint{
			Date: formatDate(row.FactDate), Population: row.Population, Eligible: row.Eligible, Known: known,
			MetricValue: ratio(row.PositiveCount, known), Coverage: ratio(known, row.Eligible),
		})
	}
	return result, nil
}

func resolveDimension(publicKey string) (models.DimensionContract, bool) {
	for _, contract := range dimensionContracts {
		if contract.PublicKey == publicKey {
			return contract, true
		}
	}
	return models.DimensionContract{}, false
}

func publicDimensionSlice(dimension models.DimensionContract, row models.DimensionRecord) models.DimensionSlice {
	value := publicDimensionValue(dimension, row.Value)
	ref := publicDimensionSliceRef(dimension, value, row.Label, row.LabelEn)
	known := row.PositiveCount + row.NegativeCount
	return models.DimensionSlice{
		Value: ref.Value, Label: ref.Label, LabelEn: ref.LabelEn,
		Population: row.Population, Eligible: row.Eligible, Known: known,
		MetricValue: ratio(row.PositiveCount, known), Coverage: ratio(known, row.Eligible),
	}
}

func publicDimensionSliceRef(dimension models.DimensionContract, value string, label, labelEn *string) models.DimensionSliceRef {
	if value == "unknown" {
		return models.DimensionSliceRef{Value: value, Label: stringPointer("未知"), LabelEn: stringPointer("Unknown")}
	}
	switch dimension.PublicKey {
	case "group":
		if label == nil {
			label = stringPointer(fmt.Sprintf("分组 #%s", value))
		}
		if labelEn == nil {
			labelEn = stringPointer(fmt.Sprintf("Group #%s", value))
		}
	case "nsfw":
		if value == "nsfw" {
			label, labelEn = stringPointer("成人内容"), stringPointer("NSFW")
		} else {
			label, labelEn = stringPointer("适合所有人"), stringPointer("SFW")
		}
	case "public_interest":
		if value == "public_interest" {
			label, labelEn = stringPointer("公益资源"), stringPointer("Public interest")
		} else {
			label, labelEn = stringPointer("常规站点"), stringPointer("Standard")
		}
	}
	return models.DimensionSliceRef{Value: value, Label: label, LabelEn: labelEn}
}

func normalizeDimensionValue(dimension models.DimensionContract, value string) (string, string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", false
	}
	switch dimension.PublicKey {
	case "country":
		if value == "unknown" {
			return value, value, true
		}
		if len(value) != 2 {
			return "", "", false
		}
		upper := strings.ToUpper(value)
		for _, r := range upper {
			if r < 'A' || r > 'Z' {
				return "", "", false
			}
		}
		return upper, upper, true
	case "group":
		if value == "unknown" {
			return value, value, true
		}
		id, err := strconv.ParseInt(value, 10, 64)
		return value, value, err == nil && id > 0
	case "nsfw":
		mappings := map[string]string{"nsfw": "true", "sfw": "false", "unknown": "unknown"}
		internal, found := mappings[value]
		return internal, value, found
	case "public_interest":
		mappings := map[string]string{"public_interest": "true", "standard": "false", "unknown": "unknown"}
		internal, found := mappings[value]
		return internal, value, found
	default:
		return "", "", false
	}
}

func publicDimensionValue(dimension models.DimensionContract, internal string) string {
	if dimension.PublicKey == "nsfw" {
		if value, ok := map[string]string{"true": "nsfw", "false": "sfw"}[internal]; ok {
			return value
		}
		return "unknown"
	}
	if dimension.PublicKey == "public_interest" {
		if value, ok := map[string]string{"true": "public_interest", "false": "standard"}[internal]; ok {
			return value
		}
		return "unknown"
	}
	return internal
}

func stringPointer(value string) *string { return &value }
