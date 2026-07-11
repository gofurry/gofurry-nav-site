package service

import (
	"fmt"
	"strings"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-center/internal/model"
)

const (
	maxNodeIDLength      = 128
	maxShortFieldLength  = 128
	maxURLLength         = 2048
	maxErrorLength       = 512
	maxErrors            = 64
	maxDisks             = 64
	maxNetworks          = 128
	maxDockerContainers  = 256
	maxChecksPerCategory = 128
)

var allowedCheckStatuses = map[string]struct{}{
	"ok":      {},
	"down":    {},
	"timeout": {},
	"warn":    {},
}

func validateAndNormalizePayload(payload *model.AgentPayload) error {
	payload.NodeID = strings.TrimSpace(payload.NodeID)
	payload.NodeName = strings.TrimSpace(payload.NodeName)
	payload.Region = strings.TrimSpace(payload.Region)
	payload.Role = strings.TrimSpace(payload.Role)
	payload.AgentVersion = strings.TrimSpace(payload.AgentVersion)
	if err := validateLength("node_id", payload.NodeID, maxNodeIDLength); err != nil {
		return err
	}
	if err := validateLength("region", payload.Region, maxShortFieldLength); err != nil {
		return err
	}
	if err := validateLength("role", payload.Role, maxShortFieldLength); err != nil {
		return err
	}
	if err := validateLength("node_name", payload.NodeName, maxShortFieldLength); err != nil {
		return err
	}
	if err := validateLength("agent_version", payload.AgentVersion, maxShortFieldLength); err != nil {
		return err
	}
	if len(payload.Errors) > maxErrors {
		return fmt.Errorf("errors exceeds %d items", maxErrors)
	}
	for i := range payload.Errors {
		payload.Errors[i] = truncateString(strings.TrimSpace(payload.Errors[i]), maxErrorLength)
	}
	if len(payload.Disks) > maxDisks {
		return fmt.Errorf("disks exceeds %d items", maxDisks)
	}
	for i := range payload.Disks {
		payload.Disks[i].Mount = strings.TrimSpace(payload.Disks[i].Mount)
		if err := validateRequiredLength("disks[].mount", payload.Disks[i].Mount, maxShortFieldLength); err != nil {
			return err
		}
	}
	if len(payload.Networks) > maxNetworks {
		return fmt.Errorf("networks exceeds %d items", maxNetworks)
	}
	for i := range payload.Networks {
		payload.Networks[i].Name = strings.TrimSpace(payload.Networks[i].Name)
		if err := validateRequiredLength("networks[].name", payload.Networks[i].Name, maxShortFieldLength); err != nil {
			return err
		}
	}
	if len(payload.Docker) > maxDockerContainers {
		return fmt.Errorf("docker exceeds %d items", maxDockerContainers)
	}
	for i := range payload.Docker {
		payload.Docker[i].Name = strings.TrimSpace(payload.Docker[i].Name)
		payload.Docker[i].Status = truncateString(strings.TrimSpace(payload.Docker[i].Status), maxShortFieldLength)
		payload.Docker[i].HealthStatus = truncateString(strings.TrimSpace(payload.Docker[i].HealthStatus), maxShortFieldLength)
		payload.Docker[i].ErrorMessage = truncateString(strings.TrimSpace(payload.Docker[i].ErrorMessage), maxErrorLength)
		if err := validateRequiredLength("docker[].name", payload.Docker[i].Name, maxShortFieldLength); err != nil {
			return err
		}
	}
	if err := normalizeHTTPChecks(payload.HTTPChecks); err != nil {
		return err
	}
	if err := normalizeServiceChecks("postgres", payload.Postgres); err != nil {
		return err
	}
	if err := normalizeServiceChecks("redis", payload.Redis); err != nil {
		return err
	}
	if err := normalizeCertChecks(payload.Certs); err != nil {
		return err
	}
	return nil
}

func normalizeHTTPChecks(items []model.HTTPCheckResult) error {
	if len(items) > maxChecksPerCategory {
		return fmt.Errorf("http_checks exceeds %d items", maxChecksPerCategory)
	}
	for i := range items {
		items[i].Name = strings.TrimSpace(items[i].Name)
		items[i].URL = strings.TrimSpace(items[i].URL)
		items[i].Status = strings.TrimSpace(items[i].Status)
		items[i].ErrorMessage = truncateString(strings.TrimSpace(items[i].ErrorMessage), maxErrorLength)
		if err := validateRequiredLength("http_checks[].name", items[i].Name, maxShortFieldLength); err != nil {
			return err
		}
		if err := validateRequiredLength("http_checks[].url", items[i].URL, maxURLLength); err != nil {
			return err
		}
		if err := validateCheckStatus("http_checks[].status", items[i].Status); err != nil {
			return err
		}
	}
	return nil
}

func normalizeServiceChecks(kind string, items []model.ServiceCheck) error {
	if len(items) > maxChecksPerCategory {
		return fmt.Errorf("%s exceeds %d items", kind, maxChecksPerCategory)
	}
	for i := range items {
		items[i].Name = strings.TrimSpace(items[i].Name)
		items[i].Status = strings.TrimSpace(items[i].Status)
		items[i].ErrorMessage = truncateString(strings.TrimSpace(items[i].ErrorMessage), maxErrorLength)
		if err := validateRequiredLength(kind+"[].name", items[i].Name, maxShortFieldLength); err != nil {
			return err
		}
		if err := validateCheckStatus(kind+"[].status", items[i].Status); err != nil {
			return err
		}
	}
	return nil
}

func normalizeCertChecks(items []model.CertCheckResult) error {
	if len(items) > maxChecksPerCategory {
		return fmt.Errorf("certs exceeds %d items", maxChecksPerCategory)
	}
	for i := range items {
		items[i].Name = strings.TrimSpace(items[i].Name)
		items[i].Host = strings.TrimSpace(items[i].Host)
		items[i].Status = strings.TrimSpace(items[i].Status)
		items[i].ErrorMessage = truncateString(strings.TrimSpace(items[i].ErrorMessage), maxErrorLength)
		if err := validateRequiredLength("certs[].name", items[i].Name, maxShortFieldLength); err != nil {
			return err
		}
		if err := validateRequiredLength("certs[].host", items[i].Host, maxShortFieldLength); err != nil {
			return err
		}
		if err := validateCheckStatus("certs[].status", items[i].Status); err != nil {
			return err
		}
	}
	return nil
}

func validateRequiredLength(name, value string, max int) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	return validateLength(name, value, max)
}

func validateLength(name, value string, max int) error {
	if len(value) > max {
		return fmt.Errorf("%s exceeds %d characters", name, max)
	}
	return nil
}

func validateCheckStatus(name, status string) error {
	if _, ok := allowedCheckStatuses[status]; !ok {
		return fmt.Errorf("%s is invalid", name)
	}
	return nil
}

func truncateString(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
