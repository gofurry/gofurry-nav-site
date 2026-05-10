package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/GoFurry/gofurry-rag/internal/db"
	"github.com/GoFurry/gofurry-rag/internal/tencentmaas"
)

const (
	minPromptBudgetRunes = 8000
	maxPromptBudgetRunes = 20000
)

type chatClient interface {
	Model() string
	Configured() bool
	Health(context.Context) error
	Complete(context.Context, []tencentmaas.Message) (tencentmaas.CompletionResult, error)
	Stream(context.Context, []tencentmaas.Message, func(string) error) (tencentmaas.CompletionResult, error)
}

type QueryCallbacks struct {
	Status  func(stage, message string) error
	Sources func(sources []db.Source) error
	Delta   func(text string) error
}

func (s *Service) StreamQuery(ctx context.Context, req QueryRequest, callbacks QueryCallbacks) (QueryResponse, error) {
	return s.executeQuery(ctx, req, callbacks)
}

func (s *Service) executeQuery(ctx context.Context, req QueryRequest, callbacks QueryCallbacks) (QueryResponse, error) {
	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		return QueryResponse{}, wrapValidation("question is required")
	}
	if limit := s.cfg.MaxQueryQuestionRunes; limit > 0 && utf8.RuneCountInString(req.Question) > limit {
		return QueryResponse{}, wrapValidation("question exceeds the maximum length")
	}

	topK := req.TopK
	if topK <= 0 {
		topK = s.cfg.TopK
	}
	if maxTopK := s.cfg.MaxQueryTopK; maxTopK > 0 && topK > maxTopK {
		return QueryResponse{}, wrapValidation("top_k exceeds the maximum limit")
	}
	if s.chat == nil || !s.chat.Configured() {
		return QueryResponse{}, wrapValidation("tencent chat model is not configured")
	}

	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.QueryTimeoutSeconds)*time.Second)
	defer cancel()

	startedAt := time.Now()
	questionRunes := utf8.RuneCountInString(req.Question)
	emitStatus := func(stage, message string) error {
		if callbacks.Status == nil {
			return nil
		}
		return callbacks.Status(stage, message)
	}
	emitSources := func(sources []db.Source) error {
		if callbacks.Sources == nil {
			return nil
		}
		return callbacks.Sources(sources)
	}

	slog.InfoContext(queryCtx, "chat query start",
		"question_runes", questionRunes,
		"top_k", topK,
		"query_timeout_seconds", s.cfg.QueryTimeoutSeconds,
		"embedding_model", s.embedder.Model(),
		"answer_model", s.chat.Model(),
	)

	if err := emitStatus("validating", "正在校验问题与参数"); err != nil {
		return QueryResponse{}, err
	}
	embeddings, err := s.embedder.Embed(queryCtx, []string{req.Question})
	if err != nil {
		slog.ErrorContext(queryCtx, "chat query embedding failed",
			"elapsed_ms", time.Since(startedAt).Milliseconds(),
			"top_k", topK,
			"embedding_model", s.embedder.Model(),
			"answer_model", s.chat.Model(),
			"error", err,
		)
		return QueryResponse{}, err
	}

	if err := emitStatus("retrieving", "正在检索相关资料"); err != nil {
		return QueryResponse{}, err
	}
	sources, err := s.repo.SearchChunks(queryCtx, embeddings[0], topK, db.BatchDocumentFilter{
		DocumentIDs: cleanDocumentIDs(req.Filters.DocumentIDs),
		SourceTypes: cleanStrings(req.Filters.SourceType),
		Categories:  cleanStrings(req.Filters.Category),
		Languages:   cleanStrings(req.Filters.Language),
	})
	if err != nil {
		slog.ErrorContext(queryCtx, "chat query search failed",
			"elapsed_ms", time.Since(startedAt).Milliseconds(),
			"top_k", topK,
			"embedding_model", s.embedder.Model(),
			"answer_model", s.chat.Model(),
			"error", err,
		)
		return QueryResponse{}, err
	}
	if err := emitSources(sources); err != nil {
		return QueryResponse{}, err
	}

	response := QueryResponse{
		Sources: sources,
		Usage: QueryUsage{
			TopK:           topK,
			EmbeddingModel: s.embedder.Model(),
			AnswerModel:    s.chat.Model(),
		},
	}
	if len(sources) == 0 {
		response.Answer = formatStructuredAnswer("当前资料中没有找到足够相关的信息。", nil, false)
		if err := emitStatus("completed", "检索完成，未找到足够资料"); err != nil {
			return QueryResponse{}, err
		}
		slog.InfoContext(queryCtx, "chat query complete",
			"elapsed_ms", time.Since(startedAt).Milliseconds(),
			"top_k", topK,
			"source_count", 0,
			"embedding_model", s.embedder.Model(),
			"answer_model", s.chat.Model(),
		)
		return response, nil
	}

	if err := emitStatus("generating", "正在生成回答"); err != nil {
		return QueryResponse{}, err
	}
	messages, usedSources := buildChatMessages(req.Question, sources, promptBudgetForTokens(s.cfg.TencentMaxTokens))

	var completion tencentmaas.CompletionResult
	if callbacks.Delta != nil {
		completion, err = s.chat.Stream(queryCtx, messages, callbacks.Delta)
	} else {
		completion, err = s.chat.Complete(queryCtx, messages)
	}
	if err != nil {
		slog.ErrorContext(queryCtx, "chat generation failed",
			"elapsed_ms", time.Since(startedAt).Milliseconds(),
			"top_k", topK,
			"source_count", len(sources),
			"embedding_model", s.embedder.Model(),
			"answer_model", s.chat.Model(),
			"error", err,
		)
		return QueryResponse{}, err
	}

	if completion.Model != "" {
		response.Usage.AnswerModel = completion.Model
	}
	if response.Usage.AnswerModel == "" {
		response.Usage.AnswerModel = s.chat.Model()
	}
	response.Answer = formatStructuredAnswer(strings.TrimSpace(completion.Answer), usedSources, len(usedSources) < len(sources))
	response.Usage.PromptTokens = completion.PromptTokens
	response.Usage.CompletionTokens = completion.CompletionTokens
	response.Usage.TotalTokens = completion.TotalTokens
	response.Usage.CachedTokens = completion.CachedTokens
	response.Usage.ReasoningTokens = completion.ReasoningTokens
	if req.IncludeDetails {
		citations, err := buildQueryCitations(queryCtx, s.repo, sources, usedSources)
		if err != nil {
			slog.WarnContext(queryCtx, "chat citation build failed",
				"elapsed_ms", time.Since(startedAt).Milliseconds(),
				"top_k", topK,
				"source_count", len(sources),
				"embedding_model", s.embedder.Model(),
				"answer_model", response.Usage.AnswerModel,
				"error", err,
			)
		} else {
			response.Citations = citations
		}
	}

	if err := emitStatus("completed", "回答已生成"); err != nil {
		return QueryResponse{}, err
	}
	slog.InfoContext(queryCtx, "chat query complete",
		"elapsed_ms", time.Since(startedAt).Milliseconds(),
		"top_k", topK,
		"source_count", len(sources),
		"used_source_count", len(usedSources),
		"embedding_model", s.embedder.Model(),
		"answer_model", response.Usage.AnswerModel,
		"prompt_tokens", response.Usage.PromptTokens,
		"completion_tokens", response.Usage.CompletionTokens,
	)
	return response, nil
}

func buildQueryCitations(ctx context.Context, repo Repository, sources, usedSources []db.Source) ([]QueryCitation, error) {
	if len(sources) == 0 {
		return nil, nil
	}

	used := make(map[int64]struct{}, len(usedSources))
	for _, source := range usedSources {
		used[source.ChunkID] = struct{}{}
	}

	docCache := make(map[int64]db.Document, len(sources))
	citations := make([]QueryCitation, 0, len(sources))
	for i, source := range sources {
		doc, ok := docCache[source.DocumentID]
		if !ok {
			fetched, err := repo.GetDocument(ctx, source.DocumentID)
			if err != nil {
				return nil, err
			}
			docCache[source.DocumentID] = fetched
			doc = fetched
		}
		_, usedInPrompt := used[source.ChunkID]
		citations = append(citations, QueryCitation{
			Rank:         i + 1,
			UsedInPrompt: usedInPrompt,
			Source:       source,
			Lineage: QueryCitationLineage{
				DocumentID: source.DocumentID,
				ChunkID:    source.ChunkID,
				ChunkIndex: source.ChunkIndex,
				SourceType: source.SourceType,
				SourceID:   source.SourceID,
				Title:      source.Title,
				URL:        source.URL,
				Score:      source.Score,
				TokenCount: source.TokenCount,
			},
			Document: buildQueryCitationDocument(doc, source),
			Chunk:    buildQueryCitationChunk(source),
		})
	}
	return citations, nil
}

func buildQueryCitationDocument(doc db.Document, source db.Source) QueryCitationDocument {
	title := doc.Title
	if title == "" {
		title = source.Title
	}
	url := doc.URL
	if url == "" {
		url = source.URL
	}
	return QueryCitationDocument{
		ID:                 doc.ID,
		SourceType:         doc.SourceType,
		SourceID:           doc.SourceID,
		Title:              title,
		URL:                url,
		Checksum:           doc.Checksum,
		Content:            doc.Content,
		Status:             doc.Status,
		ErrorMessage:       doc.ErrorMessage,
		Metadata:           doc.Metadata,
		ChunkCount:         doc.ChunkCount,
		RetryCount:         doc.RetryCount,
		LastErrorAt:        doc.LastErrorAt,
		ProcessedAt:        doc.ProcessedAt,
		ReindexRequestedAt: doc.ReindexRequestedAt,
		LastIndexedAt:      doc.LastIndexedAt,
		CreatedAt:          doc.CreatedAt,
		UpdatedAt:          doc.UpdatedAt,
	}
}

func buildQueryCitationChunk(source db.Source) QueryCitationChunk {
	return QueryCitationChunk{
		ID:          source.ChunkID,
		DocumentID:  source.DocumentID,
		ChunkIndex:  source.ChunkIndex,
		Content:     source.Content,
		ContentHash: "",
		TokenCount:  source.TokenCount,
	}
}

func buildChatMessages(question string, sources []db.Source, budgetRunes int) ([]tencentmaas.Message, []db.Source) {
	var builder strings.Builder
	budget := newRuneBudget(budgetRunes)
	usedSources := make([]db.Source, 0, len(sources))

	writeLine := func(text string) bool {
		return budget.writeLine(&builder, text)
	}

	writeLine("问题：")
	writeLine(question)
	writeLine("")
	writeLine("资料：")

	for i, source := range sources {
		if !writeLine(fmt.Sprintf("[%d] %s", i+1, trimLine(source.Title))) {
			break
		}
		if source.SourceType != "" || source.SourceID != "" {
			if !writeLine(formatSourceOrigin(source)) {
				usedSources = append(usedSources, source)
				break
			}
		}
		if source.URL != "" {
			if !writeLine("URL：" + trimLine(source.URL)) {
				usedSources = append(usedSources, source)
				break
			}
		}
		if !writeLine(fmt.Sprintf("文档ID：%d，ChunkID：%d，ChunkIndex：%d，Score：%.4f，Token：%d", source.DocumentID, source.ChunkID, source.ChunkIndex, source.Score, source.TokenCount)) {
			usedSources = append(usedSources, source)
			break
		}
		if !writeLine("内容：") {
			usedSources = append(usedSources, source)
			break
		}
		if !budget.writeText(&builder, strings.TrimSpace(source.Content)) {
			usedSources = append(usedSources, source)
			break
		}
		writeLine("")
		usedSources = append(usedSources, source)
	}

	if budget.truncated {
		writeLine("")
		writeLine("注：资料已按长度预算截断，未展示的资料不会进入模型上下文。")
	}

	systemPrompt := "你是 GoFurry RAG 控制台里的检索问答助手。\n" +
		"你只能依据我提供的资料回答，不要编造，不要补充资料外的信息。\n" +
		"请只输出答案内容，不要自行输出引用段。\n" +
		"如果资料不足，请直接回答：当前资料中没有找到足够相关的信息。"

	return []tencentmaas.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: builder.String(),
		},
	}, usedSources
}

func formatStructuredAnswer(answer string, sources []db.Source, truncated bool) string {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		answer = "当前资料中没有找到足够相关的信息。"
	}

	var builder strings.Builder
	builder.WriteString("答案：\n")
	builder.WriteString(answer)
	builder.WriteString("\n\n引用：\n")
	if len(sources) == 0 {
		builder.WriteString("无\n")
	} else {
		for i, source := range sources {
			builder.WriteString(fmt.Sprintf("- [%d] %s\n", i+1, citationSummary(source)))
		}
	}
	if truncated {
		builder.WriteString("注：引用已按长度预算截断。\n")
	}
	return strings.TrimSpace(builder.String())
}

func citationSummary(source db.Source) string {
	parts := make([]string, 0, 4)
	if title := trimLine(source.Title); title != "" {
		parts = append(parts, title)
	}
	meta := make([]string, 0, 2)
	if source.SourceType != "" {
		meta = append(meta, trimLine(source.SourceType))
	}
	if source.SourceID != "" {
		meta = append(meta, trimLine(source.SourceID))
	}
	if len(meta) > 0 {
		parts = append(parts, strings.Join(meta, "/"))
	}
	parts = append(parts, fmt.Sprintf("文档%d / Chunk%d / Score %.4f / Token %d", source.DocumentID, source.ChunkIndex, source.Score, source.TokenCount))
	if source.URL != "" {
		parts = append(parts, trimLine(source.URL))
	}
	return strings.Join(parts, " | ")
}

func formatSourceOrigin(source db.Source) string {
	if source.SourceType == "" && source.SourceID == "" {
		return ""
	}
	if source.SourceID == "" {
		return "来源：" + trimLine(source.SourceType)
	}
	if source.SourceType == "" {
		return "来源：" + trimLine(source.SourceID)
	}
	return "来源：" + trimLine(source.SourceType) + " / " + trimLine(source.SourceID)
}

func promptBudgetForTokens(maxTokens int) int {
	if maxTokens <= 0 {
		maxTokens = 1024
	}
	budget := maxTokens * 6
	if budget < minPromptBudgetRunes {
		budget = minPromptBudgetRunes
	}
	if budget > maxPromptBudgetRunes {
		budget = maxPromptBudgetRunes
	}
	return budget
}

type runeBudget struct {
	limit     int
	used      int
	truncated bool
}

func newRuneBudget(limit int) *runeBudget {
	return &runeBudget{limit: limit}
}

func (b *runeBudget) remaining() int {
	if b.limit <= 0 {
		return int(^uint(0) >> 1)
	}
	if remaining := b.limit - b.used; remaining > 0 {
		return remaining
	}
	b.truncated = true
	return 0
}

func (b *runeBudget) writeLine(builder *strings.Builder, text string) bool {
	return b.writeText(builder, text+"\n")
}

func (b *runeBudget) writeText(builder *strings.Builder, text string) bool {
	if text == "" {
		return true
	}
	remaining := b.remaining()
	if remaining <= 0 {
		return false
	}
	runes := []rune(text)
	if len(runes) <= remaining {
		builder.WriteString(text)
		b.used += len(runes)
		return true
	}
	if remaining <= 3 {
		builder.WriteString(string(runes[:remaining]))
		b.used += remaining
		b.truncated = true
		return false
	}
	keep := remaining - 3
	builder.WriteString(string(runes[:keep]))
	builder.WriteString("...")
	b.used += remaining
	b.truncated = true
	return false
}

func trimLine(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}
