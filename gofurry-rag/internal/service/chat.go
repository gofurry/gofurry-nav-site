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
		response.Answer = "当前资料中没有找到足够相关的信息。"
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
	messages := buildChatMessages(req.Question, sources)

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
	response.Answer = strings.TrimSpace(completion.Answer)
	if response.Answer == "" {
		response.Answer = "当前资料中没有找到足够相关的信息。"
	}
	response.Usage.PromptTokens = completion.PromptTokens
	response.Usage.CompletionTokens = completion.CompletionTokens
	response.Usage.TotalTokens = completion.TotalTokens
	response.Usage.CachedTokens = completion.CachedTokens
	response.Usage.ReasoningTokens = completion.ReasoningTokens

	if err := emitStatus("completed", "回答已生成"); err != nil {
		return QueryResponse{}, err
	}
	slog.InfoContext(queryCtx, "chat query complete",
		"elapsed_ms", time.Since(startedAt).Milliseconds(),
		"top_k", topK,
		"source_count", len(sources),
		"embedding_model", s.embedder.Model(),
		"answer_model", response.Usage.AnswerModel,
		"prompt_tokens", response.Usage.PromptTokens,
		"completion_tokens", response.Usage.CompletionTokens,
	)
	return response, nil
}

func buildChatMessages(question string, sources []db.Source) []tencentmaas.Message {
	var builder strings.Builder
	builder.WriteString("你是 GoFurry RAG 控制台里的知识问答助手。\n")
	builder.WriteString("你必须严格依据用户问题和检索资料作答，不要编造，不要超出资料内容。\n")
	builder.WriteString("如果资料不足以支撑回答，请直接回复：当前资料中没有找到足够相关的信息。\n")
	builder.WriteString("回答请使用简洁中文；如果需要引用资料，请使用 [1]、[2] 这样的编号标注。\n\n")
	builder.WriteString("用户问题：\n")
	builder.WriteString(question)
	builder.WriteString("\n\n检索资料：\n")
	for i, source := range sources {
		builder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, trimLine(source.Title)))
		if source.SourceType != "" || source.SourceID != "" {
			builder.WriteString("来源：")
			builder.WriteString(trimLine(source.SourceType))
			if source.SourceID != "" {
				builder.WriteString(" / ")
				builder.WriteString(trimLine(source.SourceID))
			}
			builder.WriteString("\n")
		}
		if source.URL != "" {
			builder.WriteString("URL：")
			builder.WriteString(trimLine(source.URL))
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("文档ID：%d，ChunkID：%d，ChunkIndex：%d，Score：%.4f，Token：%d\n", source.DocumentID, source.ChunkID, source.ChunkIndex, source.Score, source.TokenCount))
		builder.WriteString("内容：\n")
		builder.WriteString(strings.TrimSpace(source.Content))
		builder.WriteString("\n\n")
	}
	return []tencentmaas.Message{
		{
			Role:    "system",
			Content: "你是一个基于检索资料回答问题的助手。",
		},
		{
			Role:    "user",
			Content: builder.String(),
		},
	}
}

func trimLine(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}
