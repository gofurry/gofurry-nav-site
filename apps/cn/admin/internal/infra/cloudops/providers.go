package cloudops

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	tc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

type teoClient interface {
	CreatePurgeTaskWithContext(context.Context, *teo.CreatePurgeTaskRequest) (*teo.CreatePurgeTaskResponse, error)
	DescribePurgeTasksWithContext(context.Context, *teo.DescribePurgeTasksRequest) (*teo.DescribePurgeTasksResponse, error)
}
type edgeOne struct {
	client teoClient
	zoneID string
}

func newEdgeOne(cfg env.EdgeOneConfig) (*edgeOne, error) {
	p := profile.NewClientProfile()
	p.HttpProfile.Endpoint = "teo.tencentcloudapi.com"
	p.HttpProfile.ReqTimeout = 15
	p.NetworkFailureMaxRetries = 0
	p.RateLimitExceededMaxRetries = 0
	p.UnsafeRetryOnConnectionFailure = false
	client, err := teo.NewClient(tc.NewCredential(cfg.SecretID, cfg.SecretKey), "", p)
	if err != nil {
		return nil, errors.New("invalid EdgeOne client configuration")
	}
	return &edgeOne{client: client, zoneID: cfg.ZoneID}, nil
}
func value[T any](p *T) (result T) {
	if p != nil {
		return *p
	}
	return result
}
func pointer[T any](v T) *T { return &v }
func stringsFrom(values []*string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, value(v))
	}
	return out
}
func edgeError(err error) error {
	var apiErr *tcerr.TencentCloudSDKError
	if errors.As(err, &apiErr) {
		return fmt.Errorf("EdgeOne request failed (%s, request %s)", apiErr.GetCode(), apiErr.GetRequestId())
	}
	return errors.New("EdgeOne request failed")
}
func (e *edgeOne) purge(ctx context.Context, input PurgeRequest) (PurgeResult, error) {
	req := teo.NewCreatePurgeTaskRequest()
	req.ZoneId = &e.zoneID
	req.Type = pointer("purge_" + input.Type)
	if input.Type != "url" {
		req.Method = pointer("delete")
	}
	for _, target := range input.Targets {
		req.Targets = append(req.Targets, pointer(target))
	}
	response, err := e.client.CreatePurgeTaskWithContext(ctx, req)
	if err != nil {
		return PurgeResult{}, edgeError(err)
	}
	if response == nil || response.Response == nil {
		return PurgeResult{}, errors.New("EdgeOne returned an empty response")
	}
	data := response.Response
	result := PurgeResult{JobID: value(data.JobId), RequestID: value(data.RequestId), Status: "submitted"}
	for _, failure := range data.FailedList {
		if failure != nil {
			result.Failures = append(result.Failures, Failure{Reason: value(failure.Reason), Targets: stringsFrom(failure.Targets)})
		}
	}
	if len(result.Failures) > 0 {
		result.Status = "partial"
		if result.JobID == "" {
			result.Status = "failed"
		}
	}
	if result.JobID == "" && len(result.Failures) == 0 {
		return result, errors.New("EdgeOne did not return a task ID")
	}
	return result, nil
}
func (e *edgeOne) tasks(ctx context.Context, jobID string, offset int64) (Tasks, error) {
	req := teo.NewDescribePurgeTasksRequest()
	req.ZoneId = &e.zoneID
	req.Limit = pointer(int64(20))
	req.Offset = &offset
	if jobID != "" {
		req.Filters = []*teo.AdvancedFilter{{Name: pointer("job-id"), Values: []*string{&jobID}}}
	} else {
		now := time.Now().UTC()
		req.StartTime = pointer(now.Add(-24 * time.Hour).Format(time.RFC3339))
		req.EndTime = pointer(now.Format(time.RFC3339))
	}
	response, err := e.client.DescribePurgeTasksWithContext(ctx, req)
	if err != nil {
		return Tasks{}, edgeError(err)
	}
	if response == nil || response.Response == nil {
		return Tasks{}, errors.New("EdgeOne returned an empty response")
	}
	result := Tasks{Total: value(response.Response.TotalCount), List: []Task{}}
	for _, task := range response.Response.Tasks {
		if task != nil {
			result.List = append(result.List, Task{JobID: value(task.JobId), Target: value(task.Target), Type: strings.TrimPrefix(value(task.Type), "purge_"), Status: value(task.Status), CreatedAt: value(task.CreateTime), UpdatedAt: value(task.UpdateTime), Failure: value(task.FailType)})
		}
	}
	return result, nil
}

type cloudflare struct {
	client          *http.Client
	endpoint, token string
}

func newCloudflare(cfg env.CloudflareConfig) *cloudflare {
	return &cloudflare{client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, endpoint: "https://api.cloudflare.com/client/v4/zones/" + cfg.ZoneID + "/purge_cache", token: cfg.CacheToken}
}
func cloudflarePayload(input PurgeRequest) map[string][]string {
	targets := append([]string(nil), input.Targets...)
	field := map[string]string{"url": "files", "prefix": "prefixes", "host": "hosts"}[input.Type]
	if input.Type == "prefix" {
		for i, target := range targets {
			targets[i] = strings.TrimPrefix(strings.TrimPrefix(target, "https://"), "http://")
		}
	}
	return map[string][]string{field: targets}
}
func (c *cloudflare) purge(ctx context.Context, input PurgeRequest) (PurgeResult, error) {
	if input.Type != "url" && input.Type != "prefix" && input.Type != "host" {
		return PurgeResult{}, errors.New("unsupported Cloudflare purge type")
	}
	body, _ := json.Marshal(cloudflarePayload(input))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return PurgeResult{}, errors.New("invalid Cloudflare request")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	response, err := c.client.Do(req)
	if err != nil {
		return PurgeResult{}, errors.New("Cloudflare request failed")
	}
	defer response.Body.Close()
	var data struct {
		Success bool `json:"success"`
		Result  struct {
			ID string `json:"id"`
		} `json:"result"`
		Errors []struct {
			Code int `json:"code"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&data); err != nil {
		return PurgeResult{}, fmt.Errorf("Cloudflare response invalid (HTTP %d)", response.StatusCode)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 || !data.Success {
		code := 0
		if len(data.Errors) > 0 {
			code = data.Errors[0].Code
		}
		return PurgeResult{}, fmt.Errorf("Cloudflare purge failed (HTTP %d, code %d)", response.StatusCode, code)
	}
	if data.Result.ID == "" {
		return PurgeResult{}, errors.New("Cloudflare did not return a purge ID")
	}
	return PurgeResult{JobID: data.Result.ID, Status: "submitted"}, nil
}
