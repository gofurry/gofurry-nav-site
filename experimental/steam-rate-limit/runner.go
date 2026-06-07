package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	steam "github.com/gofurry/steam-go"
	"github.com/gofurry/steam-go/web/storefront"
	"golang.org/x/time/rate"
)

func runExperiment(ctx context.Context, cfg config) (runReport, error) {
	startedAt := time.Now()
	recorder := newEventRecorder()
	client, err := newSteamClient(cfg, recorder)
	if err != nil {
		return runReport{}, err
	}
	defer client.Close()

	cases := buildCases(cfg)
	if len(cases) == 0 {
		return runReport{}, fmt.Errorf("no request cases generated")
	}

	runner := &experimentRunner{
		cfg:      cfg,
		client:   client,
		recorder: recorder,
		cases:    cases,
		cooldown: newCooldownState(),
	}
	results, err := runner.run(ctx)
	endedAt := time.Now()

	report := runReport{
		RunID:     cfg.RunID,
		Config:    cfg.toReportConfig(),
		StartedAt: startedAt,
		EndedAt:   endedAt,
		Duration:  endedAt.Sub(startedAt),
		Results:   results,
		Events:    recorder.snapshot(),
	}
	report.Summary = summarize(report.Results)
	return report, err
}

type experimentRunner struct {
	cfg      config
	client   *steam.Client
	recorder *eventRecorder
	cases    []requestCase
	cooldown *cooldownState
	stop     atomic.Bool
}

func (r *experimentRunner) run(ctx context.Context) ([]requestResult, error) {
	jobs := make(chan requestCase)
	results := make(chan requestResult)
	var wg sync.WaitGroup
	var completed atomic.Int64
	total := int64(len(r.cases))

	fmt.Printf("开始实验: run_id=%s cases=%d tasks=%v appids=%v workers=%d burst=%d store_interval=%s api_interval=%s cooldown=%s\n",
		r.cfg.RunID, len(r.cases), r.cfg.Tasks, r.cfg.AppIDs, r.cfg.Workers, r.cfg.Burst,
		r.cfg.StoreInterval, r.cfg.APIInterval, r.cfg.CooldownOnBlock)

	for workerID := 0; workerID < r.cfg.Workers; workerID++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range jobs {
				if r.stop.Load() {
					continue
				}
				result := r.execute(ctx, item)
				results <- result
				if r.cfg.StopOnBlock && result.isBlockLike() {
					r.stop.Store(true)
				}
				if r.cfg.FailFast && !result.OK {
					r.stop.Store(true)
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, item := range r.cases {
			if r.stop.Load() {
				return
			}
			select {
			case <-ctx.Done():
				return
			case jobs <- item:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]requestResult, 0, len(r.cases))
	var ticker *time.Ticker
	if r.cfg.ProgressInterval > 0 {
		ticker = time.NewTicker(r.cfg.ProgressInterval)
		defer ticker.Stop()
	}
	for {
		select {
		case result, ok := <-results:
			if !ok {
				fmt.Printf("实验收尾: completed=%d/%d\n", completed.Load(), total)
				if err := ctx.Err(); err != nil {
					return out, err
				}
				return out, nil
			}
			out = append(out, result)
			done := completed.Add(1)
			if r.cfg.PrintEveryResult {
				fmt.Printf("[%s] done=%d/%d seq=%d task=%s appid=%d ok=%v status=%d block=%v cooldown_wait=%s duration=%s err=%s\n",
					time.Now().Format("15:04:05"), done, total, result.Seq, result.Task, result.AppID, result.OK,
					result.StatusCode, result.BlockDetected, result.CooldownWait, result.Duration, result.ErrorMessage)
			}
		case <-tickerC(ticker):
			fmt.Printf("[%s] progress=%d/%d remaining=%d store_cooldown=%s api_cooldown=%s\n",
				time.Now().Format("15:04:05"), completed.Load(), total, total-completed.Load(),
				r.cooldown.remaining(bucketStore), r.cooldown.remaining(bucketAPI))
		}
	}
}

func (r *experimentRunner) execute(ctx context.Context, item requestCase) requestResult {
	waited := r.cooldown.wait(ctx, item.Bucket)
	startedAt := time.Now()
	startSeq := r.recorder.len()
	result := requestResult{
		Seq:          item.Seq,
		Task:         item.Task,
		Bucket:       item.Bucket,
		AppID:        item.AppID,
		Region:       item.Region,
		Language:     item.Language,
		Repeat:       item.Repeat,
		StartedAt:    startedAt,
		CooldownWait: waited,
	}

	var payload []byte
	var responseSuccess *bool
	var err error

	callCtx, cancel := context.WithTimeout(ctx, r.cfg.Timeout)
	defer cancel()

	switch item.Task {
	case taskAppDetails:
		payload, err = r.client.Web.Storefront.GetAppDetailsRaw(callCtx, item.AppID, &storefront.GetAppDetailsOptions{
			CountryCode: item.Region,
			Language:    item.Language,
		})
		responseSuccess = appDetailsSuccess(item.AppID, payload)
	case taskEvents:
		payload, err = r.client.Web.Storefront.GetAdjacentPartnerEventsRaw(callCtx, item.AppID, &storefront.GetAdjacentPartnerEventsOptions{
			CountBefore:  r.cfg.EventsCountBefore,
			CountAfter:   r.cfg.EventsCountAfter,
			LanguageList: item.Language,
		})
	case taskPlayers:
		payload, err = r.client.API.SteamUserStats.GetNumberOfCurrentPlayersRaw(callCtx, item.AppID)
	default:
		err = fmt.Errorf("unsupported task: %s", item.Task)
	}

	result.EndedAt = time.Now()
	result.Duration = result.EndedAt.Sub(startedAt)
	result.ResponseBytes = len(payload)
	result.ResponseJSON = json.Valid(payload)
	result.ResponseSuccessful = responseSuccess
	if err != nil {
		result.ErrorMessage = err.Error()
	}
	event := r.recorder.lastSince(startSeq, item.Bucket)
	if event != nil {
		result.StatusCode = event.StatusCode
		result.ErrorKind = event.ErrorKind
		result.Attempts = event.Attempts
		result.CacheHit = event.CacheHit
		result.BlockDetected = event.BlockDetected
		result.EventPath = event.Path
		result.EventHost = event.Host
	}
	result.OK = err == nil && !result.BlockDetected && result.ResponseJSON
	if responseSuccess != nil && !*responseSuccess {
		result.OK = false
		result.ErrorMessage = appendError(result.ErrorMessage, "appdetails success=false")
	}
	if err != nil && result.ErrorKind == "" {
		result.ErrorKind = classifyLocalError(err)
	}
	if result.isBlockLike() && r.cfg.CooldownOnBlock > 0 {
		r.cooldown.set(item.Bucket, r.cfg.CooldownOnBlock)
	}
	return result
}

func buildCases(cfg config) []requestCase {
	var seq int64
	var cases []requestCase
	for repeat := 1; repeat <= cfg.Repeat; repeat++ {
		for _, appID := range cfg.AppIDs {
			for _, task := range cfg.Tasks {
				switch task {
				case taskAppDetails:
					for _, region := range cfg.Regions {
						for _, lang := range cfg.Languages {
							seq++
							cases = append(cases, requestCase{Seq: seq, Task: task, Bucket: bucketStore, AppID: appID, Region: region, Language: lang, Repeat: repeat})
						}
					}
				case taskEvents:
					for _, lang := range cfg.Languages {
						seq++
						cases = append(cases, requestCase{Seq: seq, Task: task, Bucket: bucketStore, AppID: appID, Language: lang, Repeat: repeat})
					}
				case taskPlayers:
					seq++
					cases = append(cases, requestCase{Seq: seq, Task: task, Bucket: bucketAPI, AppID: appID, Repeat: repeat})
				}
			}
		}
	}
	return cases
}

func newSteamClient(cfg config, recorder *eventRecorder) (*steam.Client, error) {
	backoff := steam.DefaultRetryBackoffConfig()
	backoff.BaseDelay = cfg.RetryBaseDelay
	backoff.MaxDelay = cfg.RetryBaseDelay * 4
	backoff.RespectRetryAfter = true

	options := []steam.Option{
		steam.WithTimeout(cfg.Timeout),
		steam.WithRequestObserver(steam.RequestObserverFunc(func(event steam.RequestEvent) {
			recorder.add(event)
		})),
		steam.WithTrafficPolicy(steam.TrafficClassOfficialAPI, steam.TrafficPolicy{
			RateLimiter: ratePolicy(cfg.APIInterval, cfg.Burst),
			Retry:       &steam.TrafficRetryPolicy{Retry: cfg.Retry, Backoff: backoff},
		}),
		steam.WithTrafficPolicy(steam.TrafficClassPublicStorePage, steam.TrafficPolicy{
			RateLimiter: ratePolicy(cfg.StoreInterval, cfg.Burst),
			Retry:       &steam.TrafficRetryPolicy{Retry: cfg.Retry, Backoff: backoff},
			BlockPolicy: &steam.TrafficBlockPolicy{HTMLSniffBytes: 512},
		}),
	}

	if len(cfg.ProxyURLs) > 0 {
		selector, err := steam.NewHealthCheckedRoundRobinProxySelector(steam.ProxyHealthConfig{
			FailureThreshold: 2,
			Cooldown:         cfg.CooldownOnBlock,
		}, cfg.ProxyURLs...)
		if err != nil {
			return nil, err
		}
		options = append(options, steam.WithProxySelector(selector))
	}
	return steam.NewClient(options...)
}

func ratePolicy(interval time.Duration, burst int) *steam.TrafficRateLimiterPolicy {
	if interval <= 0 || burst <= 0 {
		return nil
	}
	return &steam.TrafficRateLimiterPolicy{Limit: rate.Every(interval), Burst: burst}
}

type eventRecorder struct {
	mu     sync.RWMutex
	events []observedEvent
}

func newEventRecorder() *eventRecorder {
	return &eventRecorder{}
}

func (r *eventRecorder) add(event steam.RequestEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, observedEvent{
		TrafficClass:  string(event.TrafficClass),
		Method:        event.Method,
		Host:          event.Host,
		Path:          event.Path,
		StatusCode:    event.StatusCode,
		ErrorKind:     event.ErrorKind,
		Attempts:      event.Attempts,
		CacheHit:      event.CacheHit,
		BlockDetected: event.BlockDetected,
		Duration:      event.Duration,
		ObservedAt:    time.Now(),
	})
}

func (r *eventRecorder) len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.events)
}

func (r *eventRecorder) lastSince(start int, bucket bucketName) *observedEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for idx := len(r.events) - 1; idx >= start && idx >= 0; idx-- {
		event := r.events[idx]
		if bucketForTrafficClass(steam.TrafficClass(event.TrafficClass)) == bucket {
			return &event
		}
	}
	return nil
}

func (r *eventRecorder) snapshot() []observedEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]observedEvent, len(r.events))
	copy(out, r.events)
	return out
}

type cooldownState struct {
	mu    sync.Mutex
	until map[bucketName]time.Time
}

func newCooldownState() *cooldownState {
	return &cooldownState{until: make(map[bucketName]time.Time)}
}

func (s *cooldownState) wait(ctx context.Context, bucket bucketName) time.Duration {
	startedAt := time.Now()
	for {
		s.mu.Lock()
		until := s.until[bucket]
		s.mu.Unlock()
		wait := time.Until(until)
		if wait <= 0 {
			return time.Since(startedAt)
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return time.Since(startedAt)
		case <-timer.C:
		}
	}
}

func (s *cooldownState) set(bucket bucketName, duration time.Duration) {
	until := time.Now().Add(duration)
	s.mu.Lock()
	defer s.mu.Unlock()
	if current := s.until[bucket]; current.After(until) {
		return
	}
	s.until[bucket] = until
}

func (s *cooldownState) remaining(bucket bucketName) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	wait := time.Until(s.until[bucket])
	if wait <= 0 {
		return 0
	}
	return wait.Round(time.Second)
}

func tickerC(ticker *time.Ticker) <-chan time.Time {
	if ticker == nil {
		return nil
	}
	return ticker.C
}

func (r requestResult) isBlockLike() bool {
	return r.BlockDetected || r.StatusCode == 429 || r.StatusCode == 403 || r.StatusCode >= 500 || r.ErrorKind != ""
}

func appDetailsSuccess(appID uint32, payload []byte) *bool {
	if !json.Valid(payload) {
		return nil
	}
	var envelope map[string]struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil
	}
	item, ok := envelope[fmt.Sprintf("%d", appID)]
	if !ok {
		return nil
	}
	return &item.Success
}

func appendError(current, addition string) string {
	if current == "" {
		return addition
	}
	return current + "; " + addition
}

func classifyLocalError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	return "local"
}

func (cfg config) toReportConfig() reportConfig {
	return reportConfig{
		Tasks:             cfg.Tasks,
		AppIDs:            cfg.AppIDs,
		Regions:           cfg.Regions,
		Languages:         cfg.Languages,
		Repeat:            cfg.Repeat,
		Workers:           cfg.Workers,
		StoreInterval:     cfg.StoreInterval,
		APIInterval:       cfg.APIInterval,
		Burst:             cfg.Burst,
		Timeout:           cfg.Timeout,
		Retry:             cfg.Retry,
		RetryBaseDelay:    cfg.RetryBaseDelay,
		CooldownOnBlock:   cfg.CooldownOnBlock,
		ProgressInterval:  cfg.ProgressInterval,
		ProxyConfigured:   len(cfg.ProxyURLs) > 0,
		EventsCountBefore: cfg.EventsCountBefore,
		EventsCountAfter:  cfg.EventsCountAfter,
	}
}
