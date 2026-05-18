package repository

import (
	"strings"
	"testing"
)

func TestSchemaIncludesRawSampleIndexes(t *testing.T) {
	expected := []string{
		"idx_node_heartbeats_received",
		"idx_system_samples_reported",
		"idx_system_samples_received",
		"idx_disk_samples_reported",
		"idx_disk_samples_received",
		"idx_network_samples_node_reported",
		"idx_network_samples_reported",
		"idx_network_samples_received",
		"idx_docker_container_samples_node_reported",
		"idx_docker_container_samples_reported",
		"idx_docker_container_samples_received",
		"idx_http_check_results_node_reported",
		"idx_http_check_results_reported",
		"idx_http_check_results_received",
		"idx_service_check_results_node_reported",
		"idx_service_check_results_reported",
		"idx_service_check_results_received",
		"idx_cert_check_results_node_reported",
		"idx_cert_check_results_reported",
		"idx_cert_check_results_received",
	}
	for _, name := range expected {
		if !strings.Contains(schemaSQL, name) {
			t.Fatalf("schema missing index %s", name)
		}
	}
}
