package observation

import (
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/common/util"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

const (
	schemaVersion              = 1
	maxObservationPayloadBytes = 512 * 1024
)

func LatestKey(protocol string, siteID int64) string {
	return fmt.Sprintf("collector:v2:latest:%s:%d", protocol, siteID)
}

func SaveIfEnabled(input Input) common.GFError {
	cfg := env.GetServerConfig().Collector.V2
	if !cfg.ProtocolEnabled(input.Protocol) {
		return nil
	}
	if input.SiteID <= 0 {
		log.WarnFields(map[string]interface{}{
			"event":    "v2_observation_skipped",
			"protocol": input.Protocol,
			"reason":   "缺少 site_id",
			"target":   input.Target,
		}, "v2 observation 写入跳过：目标缺少 site_id")
		return nil
	}
	if input.ObservedAt.IsZero() {
		input.ObservedAt = time.Now()
	}
	if input.Status == "" {
		input.Status = StatusFailure
	}

	payloadBytes, err := marshalPayload(input.Payload)
	if err != nil {
		log.ErrorFields(map[string]interface{}{
			"event":    "v2_payload_encode_failed",
			"protocol": input.Protocol,
			"site_id":  input.SiteID,
			"target":   input.Target,
		}, "v2 observation payload JSON 编码失败: "+err.Error())
		return common.NewServiceError("v2 observation payload 编码失败")
	}
	if len(payloadBytes) > maxObservationPayloadBytes {
		log.ErrorFields(map[string]interface{}{
			"bytes":    len(payloadBytes),
			"event":    "v2_payload_too_large",
			"limit":    maxObservationPayloadBytes,
			"protocol": input.Protocol,
			"site_id":  input.SiteID,
			"target":   input.Target,
		}, "v2 observation payload 超过大小限制，已跳过旁路写入")
		return common.NewServiceError("v2 observation payload 超过大小限制")
	}

	var firstErr common.GFError
	if cfg.ObservationEnabled(input.Protocol) {
		record := GfnCollectorObservation{
			ID:            util.GenerateId(),
			SiteID:        input.SiteID,
			Target:        input.Target,
			Protocol:      input.Protocol,
			Status:        input.Status,
			ObservedAt:    input.ObservedAt,
			DurationMS:    input.DurationMS,
			Payload:       string(payloadBytes),
			SchemaVersion: schemaVersion,
			CreateTime:    time.Now(),
		}
		if input.ErrorCode != "" {
			record.ErrorCode = &input.ErrorCode
		}
		if input.ErrorMessage != "" {
			record.ErrorMessage = &input.ErrorMessage
		}
		if err := GetObservationDao().AddObservation(&record); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "v2_observation_db_write_failed",
				"protocol": input.Protocol,
				"site_id":  input.SiteID,
				"target":   input.Target,
			}, "v2 observation 写入数据库失败: "+err.GetMsg())
			firstErr = err
		}
	}

	if cfg.LatestRedisEnabled(input.Protocol) {
		doc := LatestDocument{
			SiteID:        input.SiteID,
			Target:        input.Target,
			Protocol:      input.Protocol,
			Status:        input.Status,
			ObservedAt:    input.ObservedAt,
			DurationMS:    input.DurationMS,
			ErrorCode:     input.ErrorCode,
			ErrorMessage:  input.ErrorMessage,
			Payload:       input.Payload,
			SchemaVersion: schemaVersion,
		}
		docBytes, err := sonic.Marshal(doc)
		if err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":    "v2_latest_encode_failed",
				"protocol": input.Protocol,
				"site_id":  input.SiteID,
				"target":   input.Target,
			}, "v2 latest Redis JSON 编码失败: "+err.Error())
			if firstErr == nil {
				firstErr = common.NewServiceError("v2 latest 编码失败")
			}
			return firstErr
		}
		key := LatestKey(input.Protocol, input.SiteID)
		if err := cs.Set(key, string(docBytes)); err != nil {
			log.ErrorFields(map[string]interface{}{
				"event":     "v2_latest_redis_write_failed",
				"protocol":  input.Protocol,
				"redis_key": key,
				"site_id":   input.SiteID,
				"target":    input.Target,
			}, "v2 latest Redis 写入失败: "+err.GetMsg())
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	if cfg.CompareLog {
		log.InfoFields(map[string]interface{}{
			"event":    "v2_observation_compared",
			"protocol": input.Protocol,
			"site_id":  input.SiteID,
			"status":   input.Status,
			"target":   input.Target,
		}, "v1/v2 采集结果旁路写入对比完成")
	}

	return firstErr
}

func marshalPayload(payload any) ([]byte, error) {
	payloadBytes, err := sonic.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if len(payloadBytes) == 0 || string(payloadBytes) == "null" {
		return []byte("{}"), nil
	}
	return payloadBytes, nil
}

func DeleteByProtocolLimit(protocol string, count string) (int64, common.GFError) {
	return GetObservationDao().DeleteByProtocolLimit(protocol, count)
}
