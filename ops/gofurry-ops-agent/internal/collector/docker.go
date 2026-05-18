package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/config"
	"github.com/gofurry/gofurry-nav-site/ops/gofurry-ops-agent/internal/model"
)

type dockerContainer struct {
	ID     string   `json:"Id"`
	Names  []string `json:"Names"`
	State  string   `json:"State"`
	Status string   `json:"Status"`
}

type dockerInspect struct {
	RestartCount int `json:"RestartCount"`
	State        struct {
		Running bool `json:"Running"`
		Health  *struct {
			Status string `json:"Status"`
		} `json:"Health"`
	} `json:"State"`
}

func collectDocker(ctx context.Context, cfg config.DockerConfig) ([]model.DockerSample, error) {
	if _, err := os.Stat(cfg.Socket); err != nil {
		return missingDockerSamples(cfg.Containers, err.Error()), err
	}
	client := dockerHTTPClient(cfg.Socket, cfg.Timeout.Duration)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/containers/json?all=1", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return missingDockerSamples(cfg.Containers, err.Error()), err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("docker api returned %s", resp.Status)
		return missingDockerSamples(cfg.Containers, err.Error()), err
	}
	var containers []dockerContainer
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return missingDockerSamples(cfg.Containers, err.Error()), err
	}

	byName := make(map[string]dockerContainer)
	for _, item := range containers {
		for _, name := range item.Names {
			name = strings.TrimPrefix(name, "/")
			byName[name] = item
		}
		if item.ID != "" {
			byName[item.ID] = item
			if len(item.ID) >= 12 {
				byName[item.ID[:12]] = item
			}
		}
	}

	if len(cfg.Containers) == 0 {
		result := make([]model.DockerSample, 0, len(containers))
		for _, item := range containers {
			result = append(result, dockerSample(ctx, client, item, firstContainerName(item)))
		}
		return result, nil
	}

	result := make([]model.DockerSample, 0, len(cfg.Containers))
	for _, name := range cfg.Containers {
		if item, ok := byName[name]; ok {
			result = append(result, dockerSample(ctx, client, item, name))
			continue
		}
		result = append(result, model.DockerSample{Name: name, Running: false, Status: "missing", ErrorMessage: "container not found"})
	}
	return result, nil
}

func dockerHTTPClient(socket string, timeout time.Duration) *http.Client {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := net.Dialer{Timeout: timeout}
			return dialer.DialContext(ctx, "unix", socket)
		},
	}
	return &http.Client{Transport: transport, Timeout: timeout}
}

func dockerSample(ctx context.Context, client *http.Client, container dockerContainer, name string) model.DockerSample {
	sample := model.DockerSample{
		Name:    name,
		ID:      container.ID,
		Running: container.State == "running",
		Status:  container.Status,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/containers/"+container.ID+"/json", nil)
	if err != nil {
		sample.ErrorMessage = err.Error()
		return sample
	}
	resp, err := client.Do(req)
	if err != nil {
		sample.ErrorMessage = err.Error()
		return sample
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		sample.ErrorMessage = "inspect returned " + resp.Status
		return sample
	}
	var inspect dockerInspect
	if err := json.NewDecoder(resp.Body).Decode(&inspect); err != nil {
		sample.ErrorMessage = err.Error()
		return sample
	}
	sample.Running = inspect.State.Running
	sample.RestartCount = inspect.RestartCount
	if inspect.State.Health != nil {
		sample.HealthStatus = inspect.State.Health.Status
	}
	return sample
}

func firstContainerName(container dockerContainer) string {
	if len(container.Names) == 0 {
		if len(container.ID) >= 12 {
			return container.ID[:12]
		}
		return container.ID
	}
	return strings.TrimPrefix(container.Names[0], "/")
}

func missingDockerSamples(names []string, message string) []model.DockerSample {
	result := make([]model.DockerSample, 0, len(names))
	for _, name := range names {
		result = append(result, model.DockerSample{Name: name, Running: false, Status: "unknown", ErrorMessage: message})
	}
	return result
}
