package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/dao"
	"github.com/gofurry/gofurry-nav-backend/apps/nav/readmodel/models"
	"github.com/gofurry/gofurry-nav-backend/common"
	cs "github.com/gofurry/gofurry-nav-backend/common/service"
)

type redisGetter func(key string) (string, common.GFError)

type observationStore interface {
	ListObservations(siteID int64, target string, protocol string, limit int) ([]models.GfnCollectorObservation, common.GFError)
}

type readModelService struct {
	get          redisGetter
	observations observationStore
}

var (
	readModelSingleton = &readModelService{}
	readModelMu        sync.Mutex
)

func GetReadModelService() *readModelService {
	readModelMu.Lock()
	defer readModelMu.Unlock()
	if readModelSingleton.get == nil {
		readModelSingleton.get = cs.GetString
	}
	if readModelSingleton.observations == nil {
		readModelSingleton.observations = dao.GetObservationDao()
	}
	return readModelSingleton
}

func newReadModelService(get redisGetter, observations observationStore) *readModelService {
	return &readModelService{get: get, observations: observations}
}

func LatestKey(protocol string, siteID int64) string {
	return fmt.Sprintf("collector:v2:latest:%s:%d", protocol, siteID)
}

func TargetLatestKey(protocol string, siteID int64, target string) string {
	return fmt.Sprintf("collector:v2:latest:%s:%d:%s", protocol, siteID, target)
}

func TargetTrendKey(siteID int64, target string) string {
	return fmt.Sprintf("collector:v2:trend:target:%d:%s", siteID, target)
}

func TargetChangeKey(siteID int64, target string) string {
	return fmt.Sprintf("collector:v2:change:target:%d:%s", siteID, target)
}

func RunStateLatestKey(protocol string) string {
	return fmt.Sprintf("collector:v2:run:%s:latest", protocol)
}

func (svc *readModelService) GetTargetLatest(siteID int64, target string, protocols []string) (models.TargetLatestResponse, common.GFError) {
	target, err := validateSiteTarget(siteID, target)
	if err != nil {
		return models.TargetLatestResponse{}, err
	}
	protocols, err = normalizeProtocols(protocols, models.CoreProtocols())
	if err != nil {
		return models.TargetLatestResponse{}, err
	}

	response := models.TargetLatestResponse{
		State:     models.SummaryStateMissing,
		SiteID:    siteID,
		Target:    target,
		Protocols: map[string]models.CollectorEnvelope{},
	}
	for _, protocol := range protocols {
		key := TargetLatestKey(protocol, siteID, target)
		raw, getErr := svc.redisGetter()(key)
		if getErr != nil {
			return models.TargetLatestResponse{}, getErr
		}
		if strings.TrimSpace(raw) == "" {
			continue
		}
		envelope, decodeErr := decodeCollectorEnvelope(raw, siteID, target, protocol)
		if decodeErr != nil {
			logRedisJSONDecodeFailure(key, siteID, target, protocol, decodeErr)
			return models.TargetLatestResponse{}, common.NewServiceError("target latest 解析失败")
		}
		response.Protocols[protocol] = envelope
	}
	if len(response.Protocols) > 0 {
		response.State = models.SummaryStateReady
	}
	return response, nil
}

func (svc *readModelService) GetLightProbeLatest(siteID int64, target string) (models.TargetLatestResponse, common.GFError) {
	return svc.GetTargetLatest(siteID, target, models.LightProbeProtocols())
}

func (svc *readModelService) ListObservations(siteID int64, target string, protocol string, limit int) (models.ObservationsResponse, common.GFError) {
	target, err := validateSiteTarget(siteID, target)
	if err != nil {
		return models.ObservationsResponse{}, err
	}
	protocol, err = validateProtocol(protocol)
	if err != nil {
		return models.ObservationsResponse{}, err
	}
	limit = models.NormalizeObservationLimit(limit)

	rows, daoErr := svc.observationStore().ListObservations(siteID, target, protocol, limit)
	if daoErr != nil {
		return models.ObservationsResponse{}, daoErr
	}
	response := models.ObservationsResponse{
		State:    models.SummaryStateMissing,
		SiteID:   siteID,
		Target:   target,
		Protocol: protocol,
		Limit:    limit,
		Items:    []models.CollectorEnvelope{},
	}
	if len(rows) == 0 {
		return response, nil
	}
	response.State = models.SummaryStateReady
	response.Items = make([]models.CollectorEnvelope, 0, len(rows))
	for _, row := range rows {
		envelope, convErr := envelopeFromObservation(row)
		if convErr != nil {
			return models.ObservationsResponse{}, common.NewServiceError("observation payload 解析失败")
		}
		response.Items = append(response.Items, envelope)
	}
	return response, nil
}

func (svc *readModelService) GetTargetTrend(siteID int64, target string) (models.TargetTrendResponse, common.GFError) {
	target, err := validateSiteTarget(siteID, target)
	if err != nil {
		return models.TargetTrendResponse{}, err
	}
	key := TargetTrendKey(siteID, target)
	raw, getErr := svc.redisGetter()(key)
	if getErr != nil {
		return models.TargetTrendResponse{}, getErr
	}
	if strings.TrimSpace(raw) == "" {
		return missingTrend(siteID, target), nil
	}
	var response models.TargetTrendResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		logRedisJSONDecodeFailure(key, siteID, target, "trend", err)
		return models.TargetTrendResponse{}, common.NewServiceError("target trend 解析失败")
	}
	response.State = models.SummaryStateReady
	if response.SiteID == 0 {
		response.SiteID = siteID
	}
	if response.Target == "" {
		response.Target = target
	}
	if len(response.Windows) == 0 {
		response.Windows = json.RawMessage(`{}`)
	}
	return response, nil
}

func (svc *readModelService) GetTargetChanges(siteID int64, target string) (models.TargetChangesResponse, common.GFError) {
	target, err := validateSiteTarget(siteID, target)
	if err != nil {
		return models.TargetChangesResponse{}, err
	}
	key := TargetChangeKey(siteID, target)
	raw, getErr := svc.redisGetter()(key)
	if getErr != nil {
		return models.TargetChangesResponse{}, getErr
	}
	if strings.TrimSpace(raw) == "" {
		return missingChanges(siteID, target), nil
	}
	var response models.TargetChangesResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		logRedisJSONDecodeFailure(key, siteID, target, "change", err)
		return models.TargetChangesResponse{}, common.NewServiceError("target changes 解析失败")
	}
	response.State = models.SummaryStateReady
	if response.SiteID == 0 {
		response.SiteID = siteID
	}
	if response.Target == "" {
		response.Target = target
	}
	if len(response.Events) == 0 {
		response.Events = json.RawMessage(`[]`)
	}
	return response, nil
}

func (svc *readModelService) GetRunState(protocol string) (models.RunStateResponse, common.GFError) {
	protocol, err := validateProtocol(protocol)
	if err != nil {
		return models.RunStateResponse{}, err
	}
	key := RunStateLatestKey(protocol)
	raw, getErr := svc.redisGetter()(key)
	if getErr != nil {
		return models.RunStateResponse{}, getErr
	}
	if strings.TrimSpace(raw) == "" {
		return models.RunStateResponse{State: models.SummaryStateMissing, Protocol: protocol}, nil
	}
	var response models.RunStateResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		logRedisJSONDecodeFailure(key, 0, "", protocol, err)
		return models.RunStateResponse{}, common.NewServiceError("run state 解析失败")
	}
	response.State = models.SummaryStateReady
	if response.Protocol == "" {
		response.Protocol = protocol
	}
	return response, nil
}

func (svc *readModelService) redisGetter() redisGetter {
	if svc != nil && svc.get != nil {
		return svc.get
	}
	return cs.GetString
}

func (svc *readModelService) observationStore() observationStore {
	if svc != nil && svc.observations != nil {
		return svc.observations
	}
	return dao.GetObservationDao()
}

func validateSiteTarget(siteID int64, target string) (string, common.GFError) {
	target = strings.TrimSpace(target)
	if siteID <= 0 {
		return "", common.NewServiceError("siteId 参数非法")
	}
	if target == "" {
		return "", common.NewServiceError("target 参数不能为空")
	}
	return target, nil
}

func validateProtocol(protocol string) (string, common.GFError) {
	protocol = strings.TrimSpace(protocol)
	if protocol == "" {
		return "", common.NewServiceError("protocol 参数不能为空")
	}
	if !models.IsProtocolAllowed(protocol) {
		return "", common.NewServiceError("protocol 参数非法")
	}
	return protocol, nil
}

func normalizeProtocols(protocols []string, defaultProtocols []string) ([]string, common.GFError) {
	if len(protocols) == 0 {
		protocols = defaultProtocols
	}
	result := make([]string, 0, len(protocols))
	seen := map[string]struct{}{}
	for _, protocol := range protocols {
		normalized, err := validateProtocol(protocol)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result, nil
}

func decodeCollectorEnvelope(raw string, siteID int64, target string, protocol string) (models.CollectorEnvelope, error) {
	var envelope models.CollectorEnvelope
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		return models.CollectorEnvelope{}, err
	}
	if envelope.SiteID == 0 {
		envelope.SiteID = siteID
	}
	if envelope.Target == "" {
		envelope.Target = target
	}
	if envelope.Protocol == "" {
		envelope.Protocol = protocol
	}
	if len(envelope.Payload) == 0 {
		envelope.Payload = json.RawMessage(`{}`)
	}
	return envelope, nil
}

func envelopeFromObservation(row models.GfnCollectorObservation) (models.CollectorEnvelope, error) {
	payload := json.RawMessage(strings.TrimSpace(row.Payload))
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	if !json.Valid(payload) {
		return models.CollectorEnvelope{}, fmt.Errorf("invalid payload")
	}
	return models.CollectorEnvelope{
		SiteID:        row.SiteID,
		Target:        row.Target,
		Protocol:      row.Protocol,
		Status:        row.Status,
		ObservedAt:    row.ObservedAt,
		DurationMS:    row.DurationMS,
		ErrorCode:     stringPtrValue(row.ErrorCode),
		ErrorMessage:  stringPtrValue(row.ErrorMessage),
		Payload:       payload,
		SchemaVersion: row.SchemaVersion,
		CollectorID:   stringFromRawPayload(payload, "collector_id"),
		JobID:         stringFromRawPayload(payload, "job_id"),
	}, nil
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringFromRawPayload(payload json.RawMessage, field string) string {
	var values map[string]json.RawMessage
	if err := json.Unmarshal(payload, &values); err != nil {
		return ""
	}
	raw, ok := values[field]
	if !ok {
		return ""
	}
	var result string
	if err := json.Unmarshal(raw, &result); err != nil {
		return ""
	}
	return result
}

func missingTrend(siteID int64, target string) models.TargetTrendResponse {
	return models.TargetTrendResponse{
		State:   models.SummaryStateMissing,
		SiteID:  siteID,
		Target:  target,
		Windows: json.RawMessage(`{}`),
	}
}

func missingChanges(siteID int64, target string) models.TargetChangesResponse {
	return models.TargetChangesResponse{
		State:  models.SummaryStateMissing,
		SiteID: siteID,
		Target: target,
		Events: json.RawMessage(`[]`),
	}
}

func logRedisJSONDecodeFailure(key string, siteID int64, target string, protocol string, err error) {
	slog.Warn(
		"collector v2 redis json parse failed",
		"redis_key", key,
		"site_id", siteID,
		"target", target,
		"protocol", protocol,
		"error", err.Error(),
	)
}
