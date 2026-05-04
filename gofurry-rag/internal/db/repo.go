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
	ID           int64           `json:"id"`
	SourceType   string          `json:"source_type"`
	SourceID     string          `json:"source_id,omitempty"`
	Title        string          `json:"title"`
	URL          string          `json:"url,omitempty"`
	Checksum     string          `json:"checksum,omitempty"`
	Content      string          `json:"-"`
	Status       string          `json:"status"`
	ErrorMessage string          `json:"error_message"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	ChunkCount   int             `json:"chunk_count"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
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
	Title      string  `json:"title"`
	URL        string  `json:"url,omitempty"`
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
	Page       int
	PageSize   int
	Status     string
	SourceType string
	Keyword    string
}

type PageResult[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

type Overview struct {
	DocumentTotal        int64      `json:"document_total"`
	ChunkTotal           int64      `json:"chunk_total"`
	EmbeddedChunkTotal   int64      `json:"embedded_chunk_total"`
	PendingDocuments     int64      `json:"pending_documents"`
	ProcessingDocuments  int64      `json:"processing_documents"`
	ReadyDocuments       int64      `json:"ready_documents"`
	FailedDocuments      int64      `json:"failed_documents"`
	LastDocumentUpdateAt *time.Time `json:"last_document_update_at,omitempty"`
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
          checksum, content, status, error_message, metadata, created_at, updated_at
`, params.SourceType, params.SourceID, params.Title, params.URL, params.Checksum, params.Content, string(params.Metadata))
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
          d.checksum, d.content, d.status, d.error_message, d.metadata, d.created_at, d.updated_at
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
	_, err = tx.Exec(ctx, `UPDATE rag_documents SET status = 'ready', error_message = '', updated_at = now() WHERE id = $1`, documentID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) MarkDocumentFailed(ctx context.Context, id int64, message string) error {
	_, err := r.pool.Exec(ctx, `UPDATE rag_documents SET status = 'failed', error_message = $2, updated_at = now() WHERE id = $1`, id, message)
	return err
}

func (r *Repository) ListDocuments(ctx context.Context, filter ListDocumentsFilter) (PageResult[Document], error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	args := []any{}
	clauses := []string{"1=1"}
	if filter.Status != "" {
		args = append(args, filter.Status)
		clauses = append(clauses, fmt.Sprintf("d.status = $%d", len(args)))
	}
	if filter.SourceType != "" {
		args = append(args, filter.SourceType)
		clauses = append(clauses, fmt.Sprintf("d.source_type = $%d", len(args)))
	}
	if filter.Keyword != "" {
		args = append(args, "%"+filter.Keyword+"%")
		clauses = append(clauses, fmt.Sprintf("(d.title ILIKE $%d OR d.id::text ILIKE $%d)", len(args), len(args)))
	}
	where := strings.Join(clauses, " AND ")

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM rag_documents d WHERE `+where, args...).Scan(&total); err != nil {
		return PageResult[Document]{}, err
	}

	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := r.pool.Query(ctx, `
SELECT d.id, d.source_type, COALESCE(d.source_id, ''), COALESCE(d.title, ''), COALESCE(d.url, ''),
       d.checksum, d.content, d.status, d.error_message, d.metadata, d.created_at, d.updated_at,
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

func (r *Repository) UpdateChunkContent(ctx context.Context, id int64, content, contentHash string, tokenCount int) (Chunk, error) {
	row := r.pool.QueryRow(ctx, `
UPDATE rag_chunks
SET content = $2, content_hash = $3, token_count = $4, embedding = NULL
WHERE id = $1
RETURNING id, document_id, chunk_index, content, content_hash, COALESCE(token_count, 0),
       embedding IS NOT NULL AS has_embedding,
       CASE WHEN embedding IS NULL THEN 0 ELSE vector_dims(embedding) END AS embedding_dim,
       created_at
`, id, content, contentHash, tokenCount)
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
    max(updated_at) AS last_document_update_at
FROM rag_documents
`).Scan(
		&overview.DocumentTotal,
		&overview.PendingDocuments,
		&overview.ProcessingDocuments,
		&overview.ReadyDocuments,
		&overview.FailedDocuments,
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
	return overview, nil
}

func (r *Repository) SearchChunks(ctx context.Context, embedding []float64, topK int) ([]Source, error) {
	if topK <= 0 {
		topK = 6
	}
	rows, err := r.pool.Query(ctx, `
SELECT c.document_id, c.id AS chunk_id, COALESCE(d.title, ''), COALESCE(d.url, ''),
       1 - (c.embedding <=> $1::vector) AS score, c.content
FROM rag_chunks c
JOIN rag_documents d ON d.id = c.document_id
WHERE d.status = 'ready' AND c.embedding IS NOT NULL
ORDER BY c.embedding <=> $1::vector
LIMIT $2
`, VectorLiteral(embedding), topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := []Source{}
	for rows.Next() {
		var item Source
		if err := rows.Scan(&item.DocumentID, &item.ChunkID, &item.Title, &item.URL, &item.Score, &item.Content); err != nil {
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
	err := row.Scan(&doc.ID, &doc.SourceType, &doc.SourceID, &doc.Title, &doc.URL, &doc.Checksum, &doc.Content, &doc.Status, &doc.ErrorMessage, &doc.Metadata, &doc.CreatedAt, &doc.UpdatedAt)
	return doc, err
}

func scanDocumentWithCount(row pgx.Row) (Document, error) {
	var doc Document
	err := row.Scan(&doc.ID, &doc.SourceType, &doc.SourceID, &doc.Title, &doc.URL, &doc.Checksum, &doc.Content, &doc.Status, &doc.ErrorMessage, &doc.Metadata, &doc.CreatedAt, &doc.UpdatedAt, &doc.ChunkCount)
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
