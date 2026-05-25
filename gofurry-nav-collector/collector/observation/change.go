package observation

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofurry/gofurry-nav-collector/common"
	"github.com/gofurry/gofurry-nav-collector/common/log"
	cs "github.com/gofurry/gofurry-nav-collector/common/service"
	"github.com/gofurry/gofurry-nav-collector/roof/env"
)

const (
	defaultChangeLookback = 60 * 24 * time.Hour
	defaultChangeTimeout  = 2 * time.Second
	defaultChangeMaxRows  = 80
	changeValueLimit      = 512
)

func TargetChangeKey(siteID int64, target string) string {
	return fmt.Sprintf("collector:v2:change:target:%d:%s", siteID, target)
}

func UpdateChangeEventsIfEnabled(siteID int64, target string, now time.Time) common.GFError {
	if siteID <= 0 || target == "" {
		return nil
	}
	cfg := env.GetServerConfig().Collector.V2
	if !cfg.Enabled || !cfg.ObservationDB || !cfg.LatestRedis {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultChangeTimeout)
	defer cancel()
	rows, err := GetObservationDao().ListChangeRows(ctx, siteID, target, now.Add(-defaultChangeLookback), defaultChangeMaxRows)
	if err != nil {
		log.WarnFields(map[string]interface{}{
			"event":   "v2_change_query_failed",
			"site_id": siteID,
			"target":  target,
		}, "v2 observation 变化事件查询失败: "+err.GetMsg())
		return err
	}
	doc := BuildTargetChanges(siteID, target, rows, now)
	bytes, marshalErr := sonic.Marshal(doc)
	if marshalErr != nil {
		log.WarnFields(map[string]interface{}{
			"event":   "v2_change_encode_failed",
			"site_id": siteID,
			"target":  target,
		}, "v2 observation 变化事件 JSON 编码失败: "+marshalErr.Error())
		return common.NewServiceError("v2 change 编码失败")
	}
	if err := cs.Set(TargetChangeKey(siteID, target), string(bytes)); err != nil {
		log.WarnFields(map[string]interface{}{
			"event":     "v2_change_redis_write_failed",
			"redis_key": TargetChangeKey(siteID, target),
			"site_id":   siteID,
			"target":    target,
		}, "v2 observation 变化事件写入 Redis 失败: "+err.GetMsg())
		return err
	}
	return nil
}

func BuildTargetChanges(siteID int64, target string, rows []ObservationTrendRow, now time.Time) TargetChangeDocument {
	grouped := map[string][]trendObservation{}
	for _, row := range rows {
		if row.Protocol != ProtocolHTTP && row.Protocol != ProtocolDNS && row.Protocol != ProtocolPortCheck && row.Protocol != ProtocolRDAP {
			continue
		}
		obs := decodeTrendObservation(row)
		if obs.Status != StatusSuccess {
			continue
		}
		grouped[row.Protocol] = append(grouped[row.Protocol], obs)
	}
	for protocol := range grouped {
		sort.Slice(grouped[protocol], func(i, j int) bool {
			return grouped[protocol][i].ObservedAt.After(grouped[protocol][j].ObservedAt)
		})
	}

	events := []ChangeEvent{}
	events = append(events, buildHTTPChangeEvents(grouped[ProtocolHTTP], now)...)
	events = append(events, buildDNSChangeEvents(grouped[ProtocolDNS], now)...)
	events = append(events, buildPortCheckChangeEvents(grouped[ProtocolPortCheck], now)...)
	events = append(events, buildRDAPChangeEvents(grouped[ProtocolRDAP], now)...)
	sort.Slice(events, func(i, j int) bool {
		if events[i].Protocol == events[j].Protocol {
			if events[i].Category == events[j].Category {
				return events[i].Field < events[j].Field
			}
			return events[i].Category < events[j].Category
		}
		return events[i].Protocol < events[j].Protocol
	})
	return TargetChangeDocument{
		SiteID:        siteID,
		Target:        target,
		Events:        events,
		GeneratedAt:   now,
		SchemaVersion: schemaVersion,
	}
}

func buildHTTPChangeEvents(values []trendObservation, now time.Time) []ChangeEvent {
	latest, previous, ok := latestPair(values)
	if !ok {
		return nil
	}
	events := []ChangeEvent{}
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "http", "status_code", normalizedNumber(latest.Payload, "status_code"), normalizedNumber(previous.Payload, "status_code"))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "http", "title", normalizedString(stringFromMap(latest.Payload, "title")), normalizedString(stringFromMap(previous.Payload, "title")))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "http", "server", normalizedString(firstNonEmptyString(stringFromMap(latest.Payload, "server"), nestedString(latest.Payload, "server_hints", "server"))), normalizedString(firstNonEmptyString(stringFromMap(previous.Payload, "server"), nestedString(previous.Payload, "server_hints", "server"))))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "http", "x_powered_by", normalizedString(nestedString(latest.Payload, "server_hints", "x_powered_by")), normalizedString(nestedString(previous.Payload, "server_hints", "x_powered_by")))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "http", "final_url", normalizedString(stringFromMap(latest.Payload, "final_url")), normalizedString(stringFromMap(previous.Payload, "final_url")))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "http", "security_headers", normalizedStringSlice(boolMapKeys(latest.Payload["security_headers"])), normalizedStringSlice(boolMapKeys(previous.Payload["security_headers"])))

	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "tls", "cert_fingerprint_sha256", normalizedString(stringFromMap(latest.Payload, "cert_fingerprint_sha256")), normalizedString(stringFromMap(previous.Payload, "cert_fingerprint_sha256")))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "tls", "cert_issuer", normalizedString(stringFromMap(latest.Payload, "cert_issuer")), normalizedString(stringFromMap(previous.Payload, "cert_issuer")))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "tls", "cert_san_count", normalizedNumber(latest.Payload, "cert_san_count"), normalizedNumber(previous.Payload, "cert_san_count"))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolHTTP, "tls", "cert_not_after", normalizedString(stringFromMap(latest.Payload, "cert_not_after")), normalizedString(stringFromMap(previous.Payload, "cert_not_after")))
	return events
}

func buildDNSChangeEvents(values []trendObservation, now time.Time) []ChangeEvent {
	latest, previous, ok := latestPair(values)
	if !ok {
		return nil
	}
	events := []ChangeEvent{}
	for _, recordType := range []string{"A", "AAAA"} {
		events = appendChangeIfDifferent(events, latest, previous, now, ProtocolDNS, "dns", recordType, normalizedStringSlice(dnsRecordValues(latest.Payload, recordType)), normalizedStringSlice(dnsRecordValues(previous.Payload, recordType)))
	}
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolDNS, "dns", "cname_terminal", normalizedString(stringFromMap(latest.Payload, "cname_terminal")), normalizedString(stringFromMap(previous.Payload, "cname_terminal")))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolDNS, "dns", "mx_hosts", normalizedStringSlice(stringSliceFromAny(latest.Payload["mx_hosts"])), normalizedStringSlice(stringSliceFromAny(previous.Payload["mx_hosts"])))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolDNS, "dns", "name_server_hosts", normalizedStringSlice(stringSliceFromAny(latest.Payload["name_server_hosts"])), normalizedStringSlice(stringSliceFromAny(previous.Payload["name_server_hosts"])))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolDNS, "dns", "soa_serial", normalizedNumber(mapFromAny(latest.Payload["soa"]), "serial"), normalizedNumber(mapFromAny(previous.Payload["soa"]), "serial"))
	return events
}

func buildPortCheckChangeEvents(values []trendObservation, now time.Time) []ChangeEvent {
	latest, previous, ok := latestPair(values)
	if !ok {
		return nil
	}
	latestPorts := portStatuses(latest.Payload)
	previousPorts := portStatuses(previous.Payload)
	ports := make([]string, 0, len(latestPorts)+len(previousPorts))
	seen := map[string]struct{}{}
	for port := range latestPorts {
		seen[port] = struct{}{}
		ports = append(ports, port)
	}
	for port := range previousPorts {
		if _, ok := seen[port]; ok {
			continue
		}
		ports = append(ports, port)
	}
	sort.Strings(ports)
	events := []ChangeEvent{}
	for _, port := range ports {
		events = appendChangeIfDifferent(events, latest, previous, now, ProtocolPortCheck, "port_check", "port_"+port, normalizedString(latestPorts[port]), normalizedString(previousPorts[port]))
	}
	return events
}

func buildRDAPChangeEvents(values []trendObservation, now time.Time) []ChangeEvent {
	latest, previous, ok := latestPair(values)
	if !ok {
		return nil
	}
	events := []ChangeEvent{}
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolRDAP, "rdap", "statuses", normalizedStringSlice(stringSliceFromAny(latest.Payload["statuses"])), normalizedStringSlice(stringSliceFromAny(previous.Payload["statuses"])))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolRDAP, "rdap", "expires_at", normalizedString(stringFromMap(latest.Payload, "expires_at")), normalizedString(stringFromMap(previous.Payload, "expires_at")))
	events = appendChangeIfDifferent(events, latest, previous, now, ProtocolRDAP, "rdap", "nameservers", normalizedStringSlice(stringSliceFromAny(latest.Payload["nameservers"])), normalizedStringSlice(stringSliceFromAny(previous.Payload["nameservers"])))
	return events
}

func latestPair(values []trendObservation) (trendObservation, trendObservation, bool) {
	if len(values) < 2 {
		return trendObservation{}, trendObservation{}, false
	}
	return values[0], values[1], true
}

type normalizedChangeValue struct {
	key   string
	value any
}

func appendChangeIfDifferent(events []ChangeEvent, latest trendObservation, previous trendObservation, now time.Time, protocol string, category string, field string, newValue normalizedChangeValue, oldValue normalizedChangeValue) []ChangeEvent {
	if newValue.key == "" && oldValue.key == "" {
		return events
	}
	if newValue.key == oldValue.key {
		return events
	}
	return append(events, ChangeEvent{
		EventID:       changeEventID(protocol, category, field, previous.ObservedAt, latest.ObservedAt),
		Protocol:      protocol,
		Category:      category,
		Field:         field,
		OldValue:      oldValue.value,
		NewValue:      newValue.value,
		OldObservedAt: previous.ObservedAt,
		NewObservedAt: latest.ObservedAt,
		DetectedAt:    now,
	})
}

func changeEventID(protocol string, category string, field string, oldObservedAt time.Time, newObservedAt time.Time) string {
	return fmt.Sprintf("%s:%s:%s:%d:%d", protocol, category, field, oldObservedAt.Unix(), newObservedAt.Unix())
}

func normalizedString(value string) normalizedChangeValue {
	value = limitChangeString(value)
	return normalizedChangeValue{key: value, value: value}
}

func normalizedNumber(values map[string]any, key string) normalizedChangeValue {
	if number, ok := floatFromMap(values, key); ok {
		text := fmt.Sprintf("%.0f", number)
		return normalizedChangeValue{key: text, value: number}
	}
	return normalizedChangeValue{}
}

func normalizedStringSlice(values []string) normalizedChangeValue {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = limitChangeString(value)
		if value != "" {
			normalized = append(normalized, value)
		}
	}
	sort.Strings(normalized)
	normalized = uniqueStrings(normalized)
	key := strings.Join(normalized, "\x00")
	return normalizedChangeValue{key: key, value: normalized}
}

func nestedString(payload map[string]any, parent string, key string) string {
	return stringFromMap(mapFromAny(payload[parent]), key)
}

func boolMapKeys(value any) []string {
	values := mapFromAny(value)
	keys := make([]string, 0, len(values))
	for key, raw := range values {
		if enabled, ok := raw.(bool); ok && enabled {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func dnsRecordValues(payload map[string]any, recordType string) []string {
	records := listFromAny(payload[recordType])
	values := make([]string, 0, len(records))
	for _, record := range records {
		if value := stringFromMap(mapFromAny(record), "value"); value != "" {
			values = append(values, value)
		}
	}
	sort.Strings(values)
	return values
}

func portStatuses(payload map[string]any) map[string]string {
	results := listFromAny(payload["results"])
	statuses := map[string]string{}
	for _, result := range results {
		item := mapFromAny(result)
		port, ok := floatFromMap(item, "port")
		if !ok {
			continue
		}
		status := stringFromMap(item, "status")
		if status == "" {
			continue
		}
		statuses[fmt.Sprintf("%.0f", port)] = status
	}
	return statuses
}

func limitChangeString(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= changeValueLimit {
		return string(runes)
	}
	return string(runes[:changeValueLimit])
}
