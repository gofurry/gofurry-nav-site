package observation

import (
	"net"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/publicsuffix"
)

const targetHintValueLimit = 2048

func BuildTargetRelationHints(target string, docs map[string]LatestDocument) (*CanonicalTargetHint, []TargetRelationHint) {
	targetHost := normalizeTargetHost(target)
	if targetHost == "" {
		return nil, nil
	}

	httpDoc, ok := docs[ProtocolHTTP]
	if !ok {
		return nil, nil
	}
	payload := mapFromAny(httpDoc.Payload)
	if len(payload) == 0 {
		return nil, nil
	}

	finalURL := limitTargetHintValue(firstNonEmptyString(
		stringFromAny(payload["final_url"]),
		stringFromAny(payload["url"]),
	))
	canonicalURL := limitTargetHintValue(firstNonEmptyString(
		stringFromAny(payload["canonical_url"]),
		stringFromAny(payload["canonicalUrl"]),
	))
	finalHost := normalizeURLHost(finalURL)
	canonicalHost := normalizeURLHost(canonicalURL)

	relations := make([]TargetRelationHint, 0, 2)
	if relation := buildTargetRelationHint(targetHost, finalHost, "final_url", finalURL); relation != nil {
		relations = append(relations, *relation)
	}
	if relation := buildTargetRelationHint(targetHost, canonicalHost, "canonical_url", canonicalURL); relation != nil {
		relations = append(relations, *relation)
	}

	preferredHost := canonicalHost
	source := "canonical_url"
	if preferredHost == "" {
		preferredHost = finalHost
		source = "final_url"
	}
	if preferredHost == "" && len(relations) == 0 {
		return nil, nil
	}

	hint := &CanonicalTargetHint{
		TargetHost:    targetHost,
		FinalHost:     finalHost,
		CanonicalHost: canonicalHost,
		PreferredHost: preferredHost,
		Relation:      hostRelation(targetHost, preferredHost),
		Source:        source,
		FinalURL:      finalURL,
		CanonicalURL:  canonicalURL,
	}
	if hint.Relation == "" {
		hint.Relation = "unknown"
	}
	return hint, relations
}

func BuildSiteTargetRelationHints(targetSummaries []TargetSummaryDocument) []SiteTargetRelationHint {
	targetHosts := map[string][]string{}
	finalHosts := map[string][]string{}
	for _, summary := range targetSummaries {
		targetHost := normalizeTargetHost(summary.Target)
		if targetHost != "" {
			targetHosts[targetHost] = append(targetHosts[targetHost], summary.Target)
		}
		if summary.CanonicalTarget != nil && summary.CanonicalTarget.FinalHost != "" {
			finalHosts[summary.CanonicalTarget.FinalHost] = append(finalHosts[summary.CanonicalTarget.FinalHost], summary.Target)
		}
	}

	hints := make([]SiteTargetRelationHint, 0)
	hints = appendDuplicateHostHints(hints, "duplicate_target_host", targetHosts)
	hints = appendDuplicateHostHints(hints, "duplicate_final_host", finalHosts)
	sort.Slice(hints, func(i, j int) bool {
		if hints[i].Relation == hints[j].Relation {
			return hints[i].Host < hints[j].Host
		}
		return hints[i].Relation < hints[j].Relation
	})
	return hints
}

func appendDuplicateHostHints(hints []SiteTargetRelationHint, relation string, hosts map[string][]string) []SiteTargetRelationHint {
	keys := make([]string, 0, len(hosts))
	for host, targets := range hosts {
		if len(uniqueStrings(targets)) > 1 {
			keys = append(keys, host)
		}
	}
	sort.Strings(keys)
	for _, host := range keys {
		targets := uniqueStrings(hosts[host])
		sort.Strings(targets)
		hints = append(hints, SiteTargetRelationHint{
			Relation: relation,
			Host:     host,
			Targets:  targets,
		})
	}
	return hints
}

func buildTargetRelationHint(targetHost string, relatedHost string, source string, value string) *TargetRelationHint {
	if targetHost == "" || relatedHost == "" {
		return nil
	}
	relation := hostRelation(targetHost, relatedHost)
	if relation == "" {
		return nil
	}
	return &TargetRelationHint{
		Relation:    relation,
		Source:      source,
		TargetHost:  targetHost,
		RelatedHost: relatedHost,
		Value:       limitTargetHintValue(value),
	}
}

func hostRelation(targetHost string, relatedHost string) string {
	targetHost = normalizeTargetHost(targetHost)
	relatedHost = normalizeTargetHost(relatedHost)
	if targetHost == "" || relatedHost == "" {
		return ""
	}
	if targetHost == relatedHost {
		return "same_host"
	}
	if relatedHost == "www."+targetHost {
		return "redirect_to_www"
	}
	if strings.HasPrefix(targetHost, "www.") && strings.TrimPrefix(targetHost, "www.") == relatedHost {
		return "redirect_to_apex"
	}
	if sameRegistrableDomain(targetHost, relatedHost) {
		return "redirect_to_related_host"
	}
	return "redirect_to_external"
}

func sameRegistrableDomain(first string, second string) bool {
	firstDomain, firstErr := publicsuffix.EffectiveTLDPlusOne(first)
	secondDomain, secondErr := publicsuffix.EffectiveTLDPlusOne(second)
	return firstErr == nil && secondErr == nil && firstDomain == secondDomain
}

func normalizeURLHost(value string) string {
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	return normalizeTargetHost(parsed.Host)
}

func normalizeTargetHost(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	if parsed, err := url.Parse(value); err == nil && parsed.Host != "" {
		value = parsed.Host
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	value = strings.Trim(value, "[]")
	value = strings.TrimSuffix(value, ".")
	return value
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func limitTargetHintValue(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= targetHintValueLimit {
		return string(runes)
	}
	return string(runes[:targetHintValueLimit])
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
