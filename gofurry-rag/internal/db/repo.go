package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusReady      = "ready"
	StatusFailed     = "failed"
)

type Repository struct {
	pool *pgxpool.Pool
}

type Document struct {
	ID                 int64           `json:"id"`
	SourceType         string          `json:"source_type"`
	SourceID           string          `json:"source_id,omitempty"`
	Title              string          `json:"title"`
	URL                string          `json:"url,omitempty"`
	Checksum           string          `json:"checksum,omitempty"`
	Content            string          `json:"-"`
	Status             string          `json:"status"`
	ErrorMessage       string          `json:"error_message"`
	Metadata           json.RawMessage `json:"metadata,omitempty"`
	ChunkCount         int             `json:"chunk_count"`
	RetryCount         int             `json:"retry_count"`
	LastErrorAt        *time.Time      `json:"last_error_at,omitempty"`
	ProcessedAt        *time.Time      `json:"processed_at,omitempty"`
	ReindexRequestedAt *time.Time      `json:"reindex_requested_at,omitempty"`
	LastIndexedAt      *time.Time      `json:"last_indexed_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type Chunk struct {
	ID           int64     `json:"id"`
	DocumentID   int64     `json:"document_id"`
	ChunkIndex   int       `json:"chunk_index"`
	Content      string    `json:"content"`
	TokenCount   int       `json:"token_count"`
	ContentHash  string    `json:"content_hash"`
	HasEmbedding bool      `json:"has_embedding"`
	EmbeddingDim int       `json:"embedding_dim"`
	CreatedAt    time.Time `json:"created_at"`
}

type Source struct {
	DocumentID int64   `json:"document_id"`
	ChunkID    int64   `json:"chunk_id"`
	SourceType string  `json:"source_type"`
	SourceID   string  `json:"source_id,omitempty"`
	Title      string  `json:"title"`
	URL        string  `json:"url,omitempty"`
	ChunkIndex int     `json:"chunk_index"`
	TokenCount int     `json:"token_count"`
	Score      float64 `json:"score"`
	Content    string  `json:"content"`
}

type CreateDocumentParams struct {
	Title      string
	Content    string
	SourceType string
	SourceID   string
	URL        string
	Checksum   string
	Metadata   json.RawMessage
}

type ListDocumentsFilter struct {
	Page        int
	PageSize    int
	Status      string
	SourceTypes []string
	Category    string
	Language    string
	Keyword     string
	DocumentIDs []int64
}

type PageResult[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

type Overview struct {
	DocumentTotal          int64      `json:"document_total"`
	ChunkTotal             int64      `json:"chunk_total"`
	EmbeddedChunkTotal     int64      `json:"embedded_chunk_total"`
	PendingDocuments       int64      `json:"pending_documents"`
	ProcessingDocuments    int64      `json:"processing_documents"`
	ReadyDocuments         int64      `json:"ready_documents"`
	FailedDocuments        int64      `json:"failed_documents"`
	QueueDocuments         int64      `json:"queue_documents"`
	RecentFailureMessage   string     `json:"recent_failure_message,omitempty"`
	RecentFailureAt        *time.Time `json:"recent_failure_at,omitempty"`
	RecentFailedDocumentID *int64     `json:"recent_failed_document_id,omitempty"`
	LastDocumentUpdateAt   *time.Time `json:"last_document_update_at,omitempty"`
}

type BatchDocumentFilter struct {
	DocumentIDs []int64
	Statuses    []string
	SourceTypes []string
	Categories  []string
	Languages   []string
}

type BatchResult struct {
	AcceptedCount int64  `json:"accepted_count"`
	SkippedCount  int64  `json:"skipped_count"`
	Status        string `json:"status"`
}

func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Migrate(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, schemaSQL)
	return err
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *Repository) CreateDocument(ctx context.Context, params CreateDocumentParams) (Document, error) {
	if len(params.Metadata) == 0 {
		params.Metadata = json.RawMessage(`{}`)
	}
	row := r.pool.QueryRow(ctx, `
INSERT INTO rag_documents (source_type, source_id, title, url, checksum, content, status, metadata)
VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7::jsonb)
RETURNING id, source_type, COALESCE(source_id, ''), COALESCE(title, ''), COALESCE(url, ''),
          checksum, content, status, error_message, metadata, retry_count, last_error_at,
          processed_at, reindex_requested_at, last_indexed_at, created_at, updated_at
`, params.SourceType, params.SourceID, params.Title, params.URL, params.Checksum, params.Content, string(params.Metadata))
	return scanDocument(row)
}

func (r *Repository) GetDocument(ctx context.Context, id int64) (Document, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, source_type, COALESCE(source_id, ''), COALESCE(title, ''), COALESCE(url, ''),
       checksum, content, status, error_message, metadata, retry_count, last_error_at,
       processed_at, reindex_requested_at, last_indexed_at, created_at, updated_at
FROM rag_documents
WHERE id = $1
`, id)
	return scanDocument(row)
}

func (r *Repository) GetDocumentByChunkID(ctx context.Context, chunkID int64) (Document, error) {
	row := r.pool.QueryRow(ctx, `
SELECT d.id, d.source_type, COALESCE(d.source_id, ''), COALESCE(d.title, ''), COALESCE(d.url, ''),
       d.checksum, d.content, d.status, d.error_message, d.metadata, d.retry_count, d.last_error_at,
       d.processed_at, d.reindex_requested_at, d.last_indexed_at, d.created_at, d.updated_at
FROM rag_documents d
JOIN rag_chunks c ON c.document_id = d.id
WHERE c.id = $1
`, chunkID)
	return scanDocument(row)
}

func (r *Repository) ClaimPendingDocument(ctx context.Context) (*Document, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
WITH picked AS (
    SELECT id FROM rag_documents
    WHERE status = 'pending'
    ORDER BY id ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
UPDATE rag_documents d
SET status = 'processing', error_message = '', updated_at = now()
FROM picked
WHERE d.id = picked.id
RETURNING d.id, d.source_type, COALESCE(d.source_id, ''), COALESCE(d.title, ''), COALESCE(d.url, ''),
          d.checksum, d.content, d.status, d.error_message, d.metadata, d.retry_count, d.last_error_at,
          d.processed_at, d.reindex_requested_at, d.last_indexed_at, d.created_at, d.updated_at
`)
	doc, err := scanDocument(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *Repository) ReplaceChunks(ctx context.Context, documentID int64, chunks []NewChunk) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM rag_chunks WHERE document_id = $1`, documentID); err != nil {
		return err
	}
	for _, chunk := range chunks {
		_, err := tx.Exec(ctx, `
INSERT INTO rag_chunks (document_id, chunk_index, content, content_hash, token_count, embedding)
VALUES ($1, $2, $3, $4, $5, $6::vector)
`, documentID, chunk.ChunkIndex, chunk.Content, chunk.ContentHash, chunk.TokenCount, VectorLiteral(chunk.Embedding))
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `
UPDATE rag_documents
SET status = 'ready',
    error_message = '',
    updated_at = now(),
    processed_at = now(),
    last_indexed_at = now()
WHERE id = $1
`, documentID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) MarkDocumentFailed(ctx context.Context, id int64, message string) error {
	_, err := r.pool.Exec(ctx, `
UPDATE rag_documents
SET status = 'failed',
    error_message = $2,
    updated_at = now(),
    last_error_at = now()
WHERE id = $1
`, id, message)
	return err
}

func (r *Repository) ReindexDocument(ctx context.Context, id int64) (Document, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Document{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM rag_chunks WHERE document_id = $1`, id); err != nil {
		return Document{}, err
	}
	row := tx.QueryRow(ctx, `
UPDATE rag_documents
SET status = 'pending',
    error_message = '',
    updated_at = now(),
    reindex_requested_at = now()
WHERE id = $1
RETURNING id, source_type, COALESCE(source_id, ''), COALESCE(title, ''), COALESCE(url, ''),
          checksum, content, status, error_message, metadata, retry_count, last_error_at,
          processed_at, reindex_requested_at, last_indexed_at, created_at, updated_at
`, id)
	doc, err := scanDocument(row)
	if err != nil {
		return Document{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Document{}, err
	}
	return doc, nil
}

func (r *Repository) ListDocuments(ctx context.Context, filter ListDocumentsFilter) (PageResult[Document], error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	documentFilter := BatchDocumentFilter{
		DocumentIDs: filter.DocumentIDs,
		SourceTypes: filter.SourceTypes,
		Categories:  singletonIfNotEmpty(filter.Category),
		Languages:   singletonIfNotEmpty(filter.Language),
	}
	if filter.Status != "" {
		documentFilter.Statuses = []string{filter.Status}
	}
	clauses, args := buildDocumentClauses(documentFilter, "d", 0)
	if filter.Keyword != "" {
		args = append(args, "%"+filter.Keyword+"%")
		placeholder := len(args)
		clauses = append(clauses, fmt.Sprintf("(d.title ILIKE $%d OR d.id::text ILIKE $%d)", placeholder, placeholder))
	}
	where := strings.Join(clauses, " AND ")

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM rag_documents d WHERE `+where, args...).Scan(&total); err != nil {
		return PageResult[Document]{}, err
	}

	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := r.pool.Query(ctx, `
SELECT d.id, d.source_type, COALESCE(d.source_id, ''), COALESCE(d.title, ''), COALESCE(d.url, ''),
       d.checksum, d.content, d.status, d.error_message, d.metadata, d.retry_count, d.last_error_at,
       d.processed_at, d.reindex_requested_at, d.last_indexed_at, d.created_at, d.updated_at,
       count(c.id)::int AS chunk_count
FROM rag_documents d
LEFT JOIN rag_chunks c ON c.document_id = d.id
WHERE `+where+`
GROUP BY d.id
ORDER BY d.id DESC
LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return PageResult[Document]{}, err
	}
	defer rows.Close()

	items := []Document{}
	for rows.Next() {
		doc, err := scanDocumentWithCount(rows)
		if err != nil {
			return PageResult[Document]{}, err
		}
		items = append(items, doc)
	}
	return PageResult[Document]{Items: items, Total: total}, rows.Err()
}

func (r *Repository) BatchReindexDocuments(ctx context.Context, filter BatchDocumentFilter) (BatchResult, error) {
	return r.requeueDocuments(ctx, filter, false)
}

func (r *Repository) RetryFailedDocuments(ctx context.Context, filter BatchDocumentFilter) (BatchResult, error) {
	return r.requeueDocuments(ctx, filter, true)
}

func (r *Repository) ListChunks(ctx context.Context, documentID int64, page, pageSize int) (PageResult[Chunk], error) {
	page, pageSize = normalizePage(page, pageSize)
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM rag_chunks WHERE document_id = $1`, documentID).Scan(&total); err != nil {
		return PageResult[Chunk]{}, err
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, document_id, chunk_index, content, content_hash, COALESCE(token_count, 0),
       embedding IS NOT NULL AS has_embedding,
       CASE WHEN embedding IS NULL THEN 0 ELSE vector_dims(embedding) END AS embedding_dim,
       created_at
FROM rag_chunks
WHERE document_id = $1
ORDER BY chunk_index ASC
LIMIT $2 OFFSET $3
`, documentID, pageSize, (page-1)*pageSize)
	if err != nil {
		return PageResult[Chunk]{}, err
	}
	defer rows.Close()

	items := []Chunk{}
	for rows.Next() {
		item, err := scanChunk(rows)
		if err != nil {
			return PageResult[Chunk]{}, err
		}
		items = append(items, item)
	}
	return PageResult[Chunk]{Items: items, Total: total}, rows.Err()
}

func (r *Repository) UpdateChunkContent(ctx context.Context, id int64, content, contentHash string, tokenCount int, embedding []float64) (Chunk, error) {
	row := r.pool.QueryRow(ctx, `
UPDATE rag_chunks
SET content = $2, content_hash = $3, token_count = $4, embedding = $5::vector
WHERE id = $1
RETURNING id, document_id, chunk_index, content, content_hash, COALESCE(token_count, 0),
       embedding IS NOT NULL AS has_embedding,
       CASE WHEN embedding IS NULL THEN 0 ELSE vector_dims(embedding) END AS embedding_dim,
       created_at
`, id, content, contentHash, tokenCount, VectorLiteral(embedding))
	return scanChunk(row)
}

func (r *Repository) DeleteChunk(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM rag_chunks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) DeleteDocument(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM rag_documents WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) Overview(ctx context.Context) (Overview, error) {
	var overview Overview
	err := r.pool.QueryRow(ctx, `
SELECT
    count(*)::bigint AS document_total,
    count(*) FILTER (WHERE status = 'pending')::bigint AS pending_documents,
    count(*) FILTER (WHERE status = 'processing')::bigint AS processing_documents,
    count(*) FILTER (WHERE status = 'ready')::bigint AS ready_documents,
    count(*) FILTER (WHERE status = 'failed')::bigint AS failed_documents,
    count(*) FILTER (WHERE status IN ('pending', 'processing'))::bigint AS queue_documents,
    max(updated_at) AS last_document_update_at
FROM rag_documents
`).Scan(
		&overview.DocumentTotal,
		&overview.PendingDocuments,
		&overview.ProcessingDocuments,
		&overview.ReadyDocuments,
		&overview.FailedDocuments,
		&overview.QueueDocuments,
		&overview.LastDocumentUpdateAt,
	)
	if err != nil {
		return Overview{}, err
	}
	err = r.pool.QueryRow(ctx, `
SELECT
    count(*)::bigint AS chunk_total,
    count(*) FILTER (WHERE embedding IS NOT NULL)::bigint AS embedded_chunk_total
FROM rag_chunks
	`).Scan(&overview.ChunkTotal, &overview.EmbeddedChunkTotal)
	if err != nil {
		return Overview{}, err
	}
	err = r.pool.QueryRow(ctx, `
SELECT id, error_message, COALESCE(last_error_at, updated_at) AS failed_at
FROM rag_documents
WHERE status = 'failed'
ORDER BY COALESCE(last_error_at, updated_at) DESC, id DESC
LIMIT 1
`).Scan(&overview.RecentFailedDocumentID, &overview.RecentFailureMessage, &overview.RecentFailureAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Overview{}, err
	}
	return overview, nil
}

func (r *Repository) SearchChunks(ctx context.Context, embedding []float64, topK int, filter BatchDocumentFilter) ([]Source, error) {
	if topK <= 0 {
		topK = 6
	}
	filter.Statuses = []string{StatusReady}
	clauses, args := buildDocumentClauses(filter, "d", 1)
	args = append(args, topK)
	limitPlaceholder := 1 + len(args)
	rows, err := r.pool.Query(ctx, `
SELECT c.document_id, c.id AS chunk_id, d.source_type, COALESCE(d.source_id, ''),
       COALESCE(d.title, ''), COALESCE(d.url, ''),
       c.chunk_index, COALESCE(c.token_count, 0),
       1 - (c.embedding <=> $1::vector) AS score, c.content
FROM rag_chunks c
JOIN rag_documents d ON d.id = c.document_id
WHERE c.embedding IS NOT NULL AND `+strings.Join(clauses, " AND ")+`
ORDER BY c.embedding <=> $1::vector
LIMIT $`+strconv.Itoa(limitPlaceholder)+`
`, append([]any{VectorLiteral(embedding)}, args...)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []Source{}
	for rows.Next() {
		var item Source
		if err := rows.Scan(&item.DocumentID, &item.ChunkID, &item.SourceType, &item.SourceID, &item.Title, &item.URL, &item.ChunkIndex, &item.TokenCount, &item.Score, &item.Content); err != nil {
			return nil, err
		}
		sources = append(sources, item)
	}
	return sources, rows.Err()
}

type NewChunk struct {
	ChunkIndex  int
	Content     string
	ContentHash string
	TokenCount  int
	Embedding   []float64
}

func VectorLiteral(values []float64) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			value = 0
		}
		parts = append(parts, strconv.FormatFloat(value, 'g', -1, 64))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func scanDocument(row pgx.Row) (Document, error) {
	var doc Document
	err := row.Scan(
		&doc.ID,
		&doc.SourceType,
		&doc.SourceID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.Content,
		&doc.Status,
		&doc.ErrorMessage,
		&doc.Metadata,
		&doc.RetryCount,
		&doc.LastErrorAt,
		&doc.ProcessedAt,
		&doc.ReindexRequestedAt,
		&doc.LastIndexedAt,
		&doc.CreatedAt,
		&doc.UpdatedAt,
	)
	return doc, err
}

func scanDocumentWithCount(row pgx.Row) (Document, error) {
	var doc Document
	err := row.Scan(
		&doc.ID,
		&doc.SourceType,
		&doc.SourceID,
		&doc.Title,
		&doc.URL,
		&doc.Checksum,
		&doc.Content,
		&doc.Status,
		&doc.ErrorMessage,
		&doc.Metadata,
		&doc.RetryCount,
		&doc.LastErrorAt,
		&doc.ProcessedAt,
		&doc.ReindexRequestedAt,
		&doc.LastIndexedAt,
		&doc.CreatedAt,
		&doc.UpdatedAt,
		&doc.ChunkCount,
	)
	return doc, err
}

func scanChunk(row pgx.Row) (Chunk, error) {
	var item Chunk
	err := row.Scan(&item.ID, &item.DocumentID, &item.ChunkIndex, &item.Content, &item.ContentHash, &item.TokenCount, &item.HasEmbedding, &item.EmbeddingDim, &item.CreatedAt)
	return item, err
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (r *Repository) requeueDocuments(ctx context.Context, filter BatchDocumentFilter, failedOnly bool) (BatchResult, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return BatchResult{}, err
	}
	defer tx.Rollback(ctx)

	baseFilter := filter
	total, err := r.countDocumentsTx(ctx, tx, baseFilter)
	if err != nil {
		return BatchResult{}, err
	}

	eligibleFilter := filter
	if failedOnly {
		eligibleFilter.Statuses = failedOnlyStatuses(eligibleFilter.Statuses)
	} else if len(eligibleFilter.Statuses) == 0 {
		eligibleFilter.Statuses = []string{StatusPending, StatusProcessing, StatusReady, StatusFailed}
	}
	clauses, args := buildDocumentClauses(eligibleFilter, "d", 0)
	where := strings.Join(clauses, " AND ")

	rows, err := tx.Query(ctx, `SELECT d.id FROM rag_documents d WHERE `+where, args...)
	if err != nil {
		return BatchResult{}, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return BatchResult{}, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return BatchResult{}, err
	}
	if len(ids) == 0 {
		return BatchResult{AcceptedCount: 0, SkippedCount: total, Status: StatusPending}, tx.Commit(ctx)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM rag_chunks WHERE document_id = ANY($1)`, ids); err != nil {
		return BatchResult{}, err
	}

	query := `
UPDATE rag_documents
SET status = 'pending',
    error_message = '',
    updated_at = now(),
    reindex_requested_at = now()
`
	if failedOnly {
		query += `,
    retry_count = retry_count + 1`
	}
	query += `
WHERE id = ANY($1)
`
	tag, err := tx.Exec(ctx, query, ids)
	if err != nil {
		return BatchResult{}, err
	}

	result := BatchResult{
		AcceptedCount: tag.RowsAffected(),
		SkippedCount:  total - tag.RowsAffected(),
		Status:        StatusPending,
	}
	if err := tx.Commit(ctx); err != nil {
		return BatchResult{}, err
	}
	return result, nil
}

func (r *Repository) countDocumentsTx(ctx context.Context, tx pgx.Tx, filter BatchDocumentFilter) (int64, error) {
	clauses, args := buildDocumentClauses(filter, "d", 0)
	var total int64
	err := tx.QueryRow(ctx, `SELECT count(*) FROM rag_documents d WHERE `+strings.Join(clauses, " AND "), args...).Scan(&total)
	return total, err
}

func buildDocumentClauses(filter BatchDocumentFilter, alias string, start int) ([]string, []any) {
	clauses := []string{"1=1"}
	args := make([]any, 0)
	addArrayClause := func(values any, expr string) {
		switch typed := values.(type) {
		case []string:
			if len(typed) == 0 {
				return
			}
		case []int64:
			if len(typed) == 0 {
				return
			}
		}
		args = append(args, values)
		clauses = append(clauses, fmt.Sprintf(expr, start+len(args)))
	}

	addArrayClause(nonEmptyStrings(filter.Statuses), alias+".status = ANY($%d)")
	addArrayClause(filter.DocumentIDs, alias+".id = ANY($%d)")
	addArrayClause(nonEmptyStrings(filter.SourceTypes), alias+".source_type = ANY($%d)")
	addArrayClause(nonEmptyStrings(filter.Categories), "COALESCE("+alias+".metadata->>'category', '') = ANY($%d)")
	addArrayClause(nonEmptyStrings(filter.Languages), "COALESCE("+alias+".metadata->>'language', '') = ANY($%d)")
	return clauses, args
}

func nonEmptyStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}

func singletonIfNotEmpty(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return []string{value}
}

func failedOnlyStatuses(statuses []string) []string {
	if len(statuses) == 0 {
		return []string{StatusFailed}
	}
	for _, status := range statuses {
		if strings.TrimSpace(status) == StatusFailed {
			return []string{StatusFailed}
		}
	}
	return []string{"__no_match__"}
}
