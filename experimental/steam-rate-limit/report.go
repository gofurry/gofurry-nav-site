package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

func writeOutputs(report runReport, outputDir string) ([]string, error) {
	dir := filepath.Join(outputDir, report.RunID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	jsonPath := filepath.Join(dir, "report.json")
	csvPath := filepath.Join(dir, "results.csv")
	mdPath := filepath.Join(dir, "report.zh-CN.md")

	report.Generated = []string{jsonPath, csvPath, mdPath}

	if err := writeJSON(jsonPath, report); err != nil {
		return nil, err
	}
	if err := writeCSV(csvPath, report.Results); err != nil {
		return nil, err
	}
	if err := os.WriteFile(mdPath, []byte(renderChineseReport(report)), 0644); err != nil {
		return nil, err
	}
	return report.Generated, nil
}

func writeJSON(path string, report runReport) error {
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0644)
}

func writeCSV(path string, results []requestResult) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"seq", "task", "bucket", "appid", "region", "language", "repeat",
		"started_at", "duration_ms", "ok", "status_code", "error_kind",
		"block_detected", "attempts", "response_bytes", "response_json",
		"response_successful", "cooldown_wait_ms", "event_host", "event_path", "error_message",
	}
	if err := writer.Write(header); err != nil {
		return err
	}
	for _, item := range results {
		responseSuccessful := ""
		if item.ResponseSuccessful != nil {
			responseSuccessful = strconv.FormatBool(*item.ResponseSuccessful)
		}
		row := []string{
			strconv.FormatInt(item.Seq, 10),
			string(item.Task),
			string(item.Bucket),
			strconv.FormatUint(uint64(item.AppID), 10),
			item.Region,
			item.Language,
			strconv.Itoa(item.Repeat),
			item.StartedAt.Format(time.RFC3339),
			strconv.FormatInt(item.Duration.Milliseconds(), 10),
			strconv.FormatBool(item.OK),
			strconv.Itoa(item.StatusCode),
			item.ErrorKind,
			strconv.FormatBool(item.BlockDetected),
			strconv.Itoa(item.Attempts),
			strconv.Itoa(item.ResponseBytes),
			strconv.FormatBool(item.ResponseJSON),
			responseSuccessful,
			strconv.FormatInt(item.CooldownWait.Milliseconds(), 10),
			item.EventHost,
			item.EventPath,
			item.ErrorMessage,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return writer.Error()
}

func summarize(results []requestResult) reportSummary {
	summary := reportSummary{
		Total:    len(results),
		ByTask:   make(map[taskName]taskSummary),
		ByBucket: make(map[bucketName]taskSummary),
	}
	for _, item := range results {
		if item.OK {
			summary.OK++
		} else {
			summary.Failed++
		}
		if item.BlockDetected {
			summary.Blocked++
		}
		switch {
		case item.StatusCode == 429:
			summary.HTTP429++
		case item.StatusCode == 403:
			summary.HTTP403++
		case item.StatusCode >= 500:
			summary.HTTP5xx++
		}
		if item.ErrorKind == "transport" || item.ErrorKind == "timeout" {
			summary.TimeoutOrTransport++
		}
		summary.ByTask[item.Task] = addSummary(summary.ByTask[item.Task], item)
		summary.ByBucket[item.Bucket] = addSummary(summary.ByBucket[item.Bucket], item)
	}
	summary.ByTask = finalizeSummaries(summary.ByTask, results, func(item requestResult) taskName { return item.Task })
	summary.ByBucket = finalizeSummaries(summary.ByBucket, results, func(item requestResult) bucketName { return item.Bucket })
	summary.Recommendation = buildRecommendation(summary)
	return summary
}

func addSummary(summary taskSummary, item requestResult) taskSummary {
	summary.Total++
	if item.OK {
		summary.OK++
	} else {
		summary.Failed++
	}
	if item.BlockDetected {
		summary.Blocked++
	}
	return summary
}

func finalizeSummaries[K comparable](summaries map[K]taskSummary, results []requestResult, key func(requestResult) K) map[K]taskSummary {
	durations := make(map[K][]time.Duration)
	for _, item := range results {
		k := key(item)
		durations[k] = append(durations[k], item.Duration)
	}
	for k, values := range durations {
		sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
		var total time.Duration
		for _, value := range values {
			total += value
		}
		summary := summaries[k]
		if len(values) > 0 {
			summary.AvgDuration = total / time.Duration(len(values))
			summary.P95Duration = values[p95Index(len(values))]
		}
		summaries[k] = summary
	}
	return summaries
}

func p95Index(length int) int {
	if length <= 1 {
		return 0
	}
	idx := int(float64(length)*0.95) - 1
	if idx < 0 {
		return 0
	}
	if idx >= length {
		return length - 1
	}
	return idx
}

func buildRecommendation(summary reportSummary) string {
	if summary.Total == 0 {
		return "没有有效样本，无法给出建议。"
	}
	if summary.HTTP429 > 0 || summary.Blocked > 0 {
		return "本轮出现 429 或 block detection，建议下调并发/提高请求间隔，并保留 cooldown。"
	}
	if summary.HTTP403 > 0 {
		return "本轮出现 403，Store 类请求不应提高并发；优先检查代理、地区和请求头环境。"
	}
	if summary.HTTP5xx > 0 || summary.TimeoutOrTransport > 0 {
		return "本轮存在 5xx/timeout/transport 错误，建议重复实验确认是否为网络波动，再调整生产默认值。"
	}
	if summary.Failed > 0 {
		return "本轮存在非风控失败，建议先查看 CSV 中的 error_message，再决定是否调整限流。"
	}
	return "本轮未观察到明显风控信号，可以在下一轮小幅降低间隔或提高样本量继续验证。"
}

func renderChineseReport(report runReport) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Steam 限流规则校准实验报告\n\n")
	fmt.Fprintf(&buf, "- Run ID：`%s`\n", report.RunID)
	fmt.Fprintf(&buf, "- 开始时间：`%s`\n", report.StartedAt.Format(time.RFC3339))
	fmt.Fprintf(&buf, "- 结束时间：`%s`\n", report.EndedAt.Format(time.RFC3339))
	fmt.Fprintf(&buf, "- 总耗时：`%s`\n", report.Duration)
	fmt.Fprintf(&buf, "- 样本数：`%d`\n\n", report.Summary.Total)

	fmt.Fprintf(&buf, "## 实验参数\n\n")
	fmt.Fprintf(&buf, "| 参数 | 值 |\n| --- | --- |\n")
	fmt.Fprintf(&buf, "| tasks | `%v` |\n", report.Config.Tasks)
	fmt.Fprintf(&buf, "| appids | `%v` |\n", report.Config.AppIDs)
	fmt.Fprintf(&buf, "| regions | `%v` |\n", report.Config.Regions)
	fmt.Fprintf(&buf, "| languages | `%v` |\n", report.Config.Languages)
	fmt.Fprintf(&buf, "| repeat | `%d` |\n", report.Config.Repeat)
	fmt.Fprintf(&buf, "| workers | `%d` |\n", report.Config.Workers)
	fmt.Fprintf(&buf, "| store_interval | `%s` |\n", report.Config.StoreInterval)
	fmt.Fprintf(&buf, "| api_interval | `%s` |\n", report.Config.APIInterval)
	fmt.Fprintf(&buf, "| burst | `%d` |\n", report.Config.Burst)
	fmt.Fprintf(&buf, "| timeout | `%s` |\n", report.Config.Timeout)
	fmt.Fprintf(&buf, "| retry | `%d` |\n", report.Config.Retry)
	fmt.Fprintf(&buf, "| retry_base_delay | `%s` |\n", report.Config.RetryBaseDelay)
	fmt.Fprintf(&buf, "| cooldown_on_block | `%s` |\n", report.Config.CooldownOnBlock)
	fmt.Fprintf(&buf, "| progress_interval | `%s` |\n", report.Config.ProgressInterval)
	fmt.Fprintf(&buf, "| proxy_configured | `%v` |\n\n", report.Config.ProxyConfigured)

	fmt.Fprintf(&buf, "## 总览\n\n")
	fmt.Fprintf(&buf, "| 指标 | 值 |\n| --- | ---: |\n")
	fmt.Fprintf(&buf, "| total | %d |\n", report.Summary.Total)
	fmt.Fprintf(&buf, "| ok | %d |\n", report.Summary.OK)
	fmt.Fprintf(&buf, "| failed | %d |\n", report.Summary.Failed)
	fmt.Fprintf(&buf, "| blocked | %d |\n", report.Summary.Blocked)
	fmt.Fprintf(&buf, "| HTTP 429 | %d |\n", report.Summary.HTTP429)
	fmt.Fprintf(&buf, "| HTTP 403 | %d |\n", report.Summary.HTTP403)
	fmt.Fprintf(&buf, "| HTTP 5xx | %d |\n", report.Summary.HTTP5xx)
	fmt.Fprintf(&buf, "| timeout/transport | %d |\n\n", report.Summary.TimeoutOrTransport)

	fmt.Fprintf(&buf, "## 按任务统计\n\n")
	renderTaskTable(&buf, report.Summary.ByTask)
	fmt.Fprintf(&buf, "\n## 按流量桶统计\n\n")
	renderBucketTable(&buf, report.Summary.ByBucket)

	fmt.Fprintf(&buf, "\n## 结论建议\n\n")
	fmt.Fprintf(&buf, "%s\n\n", report.Summary.Recommendation)
	fmt.Fprintf(&buf, "## 生产配置判断口径\n\n")
	fmt.Fprintf(&buf, "- 若出现 `429`、`403` 或 `block_detected=true`，本轮配置不能作为生产默认值。\n")
	fmt.Fprintf(&buf, "- 若 Store 任务失败集中在 `appdetails`，优先提高 `store-interval-ms`，不要直接加 workers。\n")
	fmt.Fprintf(&buf, "- 若 official API 任务出现 `429`，优先提高 `api-interval-ms`，并保持独立于 Store bucket。\n")
	fmt.Fprintf(&buf, "- 建议每次只改变一个变量：interval、burst、workers、proxy，避免无法归因。\n")
	fmt.Fprintf(&buf, "- 当前经验规则仍应作为下限保护：official API 约 `100 token / 1 minute`，Store 约 `[150, 250] token / 5 minutes`。\n\n")
	fmt.Fprintf(&buf, "## 输出文件\n\n")
	for _, path := range report.Generated {
		fmt.Fprintf(&buf, "- `%s`\n", path)
	}
	return buf.String()
}

func renderTaskTable(buf *bytes.Buffer, summaries map[taskName]taskSummary) {
	fmt.Fprintf(buf, "| 任务 | total | ok | failed | blocked | avg | p95 |\n| --- | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	keys := make([]string, 0, len(summaries))
	for key := range summaries {
		keys = append(keys, string(key))
	}
	sort.Strings(keys)
	for _, raw := range keys {
		summary := summaries[taskName(raw)]
		fmt.Fprintf(buf, "| `%s` | %d | %d | %d | %d | `%s` | `%s` |\n", raw, summary.Total, summary.OK, summary.Failed, summary.Blocked, summary.AvgDuration, summary.P95Duration)
	}
}

func renderBucketTable(buf *bytes.Buffer, summaries map[bucketName]taskSummary) {
	fmt.Fprintf(buf, "| bucket | total | ok | failed | blocked | avg | p95 |\n| --- | ---: | ---: | ---: | ---: | ---: | ---: |\n")
	keys := make([]string, 0, len(summaries))
	for key := range summaries {
		keys = append(keys, string(key))
	}
	sort.Strings(keys)
	for _, raw := range keys {
		summary := summaries[bucketName(raw)]
		fmt.Fprintf(buf, "| `%s` | %d | %d | %d | %d | `%s` | `%s` |\n", raw, summary.Total, summary.OK, summary.Failed, summary.Blocked, summary.AvgDuration, summary.P95Duration)
	}
}
