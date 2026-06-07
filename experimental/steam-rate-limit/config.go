package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func parseConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("steam-rate-limit", flag.ContinueOnError)

	var rawTasks, rawAppIDs, appIDFile, rawRegions, rawLanguages, rawProxy string
	var storeIntervalMS, apiIntervalMS, timeoutSeconds, retryBaseMS, cooldownSeconds, progressSeconds int

	cfg := config{}
	fs.StringVar(&cfg.RunID, "run-id", "", "run id; default uses current timestamp")
	fs.StringVar(&rawTasks, "tasks", "appdetails,events,players", "comma-separated tasks: appdetails,events,players")
	fs.StringVar(&rawAppIDs, "appids", "440,570,730", "comma-separated Steam appids")
	fs.StringVar(&appIDFile, "appid-file", "", "optional file with one appid per line")
	fs.StringVar(&rawRegions, "regions", "CN,US,HK", "comma-separated Store country codes for appdetails")
	fs.StringVar(&rawLanguages, "languages", "schinese,english", "comma-separated Steam language names")
	fs.IntVar(&cfg.Repeat, "repeat", 1, "repeat count for every generated request case")
	fs.IntVar(&cfg.Workers, "workers", 1, "concurrent workers")
	fs.IntVar(&storeIntervalMS, "store-interval-ms", 2000, "Store traffic token interval in milliseconds")
	fs.IntVar(&apiIntervalMS, "api-interval-ms", 1000, "official API traffic token interval in milliseconds")
	fs.IntVar(&cfg.Burst, "burst", 1, "token bucket burst")
	fs.IntVar(&timeoutSeconds, "timeout-seconds", 15, "single request timeout in seconds")
	fs.IntVar(&cfg.Retry, "retry", 1, "SDK retry count")
	fs.IntVar(&retryBaseMS, "retry-base-ms", 1000, "SDK retry base delay in milliseconds")
	fs.IntVar(&cooldownSeconds, "cooldown-on-block-seconds", 300, "experiment cooldown after 403/429/5xx/error/block")
	fs.IntVar(&progressSeconds, "progress-seconds", 5, "progress log interval in seconds; 0 disables progress logs")
	fs.StringVar(&rawProxy, "proxy", "", "comma/semicolon/newline separated proxy URLs")
	fs.IntVar(&cfg.EventsCountBefore, "events-count-before", 1, "Steam events count_before")
	fs.IntVar(&cfg.EventsCountAfter, "events-count-after", 10, "Steam events count_after")
	fs.StringVar(&cfg.OutputDir, "out", "out", "output directory")
	fs.BoolVar(&cfg.StopOnBlock, "stop-on-block", false, "stop scheduling new work when block/rate-limit is observed")
	fs.BoolVar(&cfg.FailFast, "fail-fast", false, "stop on the first logical request error")
	fs.BoolVar(&cfg.PrintEveryResult, "print-each", false, "print every logical request result")

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}

	if cfg.RunID == "" {
		cfg.RunID = time.Now().Format("20060102-150405")
	}
	tasks, err := parseTasks(rawTasks)
	if err != nil {
		return config{}, err
	}
	cfg.Tasks = tasks

	appIDs, err := parseAppIDs(rawAppIDs)
	if err != nil {
		return config{}, err
	}
	if appIDFile != "" {
		fileAppIDs, err := readAppIDFile(appIDFile)
		if err != nil {
			return config{}, err
		}
		appIDs = append(appIDs, fileAppIDs...)
	}
	cfg.AppIDs = uniqueAppIDs(appIDs)
	if len(cfg.AppIDs) == 0 {
		return config{}, fmt.Errorf("at least one appid is required")
	}

	cfg.Regions = parseCSV(rawRegions)
	cfg.Languages = parseCSV(rawLanguages)
	cfg.ProxyURLs = parseCSV(rawProxy)
	cfg.StoreInterval = time.Duration(storeIntervalMS) * time.Millisecond
	cfg.APIInterval = time.Duration(apiIntervalMS) * time.Millisecond
	cfg.Timeout = time.Duration(timeoutSeconds) * time.Second
	cfg.RetryBaseDelay = time.Duration(retryBaseMS) * time.Millisecond
	cfg.CooldownOnBlock = time.Duration(cooldownSeconds) * time.Second
	cfg.ProgressInterval = time.Duration(progressSeconds) * time.Second

	if err := validateConfig(cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func validateConfig(cfg config) error {
	if cfg.Repeat <= 0 {
		return fmt.Errorf("repeat must be greater than zero")
	}
	if cfg.Workers <= 0 {
		return fmt.Errorf("workers must be greater than zero")
	}
	if cfg.StoreInterval < 0 || cfg.APIInterval < 0 {
		return fmt.Errorf("intervals must not be negative")
	}
	if cfg.Burst < 0 {
		return fmt.Errorf("burst must not be negative")
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than zero")
	}
	if cfg.Retry < 0 {
		return fmt.Errorf("retry must not be negative")
	}
	if cfg.RetryBaseDelay <= 0 {
		return fmt.Errorf("retry base delay must be greater than zero")
	}
	if cfg.CooldownOnBlock < 0 {
		return fmt.Errorf("cooldown must not be negative")
	}
	if cfg.ProgressInterval < 0 {
		return fmt.Errorf("progress interval must not be negative")
	}
	if containsTask(cfg.Tasks, taskAppDetails) && len(cfg.Regions) == 0 {
		return fmt.Errorf("regions are required for appdetails")
	}
	if (containsTask(cfg.Tasks, taskAppDetails) || containsTask(cfg.Tasks, taskEvents)) && len(cfg.Languages) == 0 {
		return fmt.Errorf("languages are required for Store tasks")
	}
	return nil
}

func parseTasks(raw string) ([]taskName, error) {
	parts := parseCSV(raw)
	tasks := make([]taskName, 0, len(parts))
	for _, part := range parts {
		task := taskName(strings.ToLower(part))
		switch task {
		case taskAppDetails, taskEvents, taskPlayers:
			tasks = append(tasks, task)
		default:
			return nil, fmt.Errorf("unsupported task: %s", part)
		}
	}
	return tasks, nil
}

func parseCSV(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseAppIDs(raw string) ([]uint32, error) {
	parts := parseCSV(raw)
	out := make([]uint32, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseUint(part, 10, 32)
		if err != nil || value == 0 {
			return nil, fmt.Errorf("invalid appid: %s", part)
		}
		out = append(out, uint32(value))
	}
	return out, nil
}

func readAppIDFile(path string) ([]uint32, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ids []uint32
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parsed, err := parseAppIDs(line)
		if err != nil {
			return nil, err
		}
		ids = append(ids, parsed...)
	}
	return ids, scanner.Err()
}

func uniqueAppIDs(ids []uint32) []uint32 {
	seen := make(map[uint32]struct{}, len(ids))
	out := make([]uint32, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func containsTask(tasks []taskName, want taskName) bool {
	for _, task := range tasks {
		if task == want {
			return true
		}
	}
	return false
}
