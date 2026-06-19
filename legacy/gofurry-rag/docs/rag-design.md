# gofurry-rag MVP Design

The canonical design draft is currently tracked at the repository root as `gofurry-rag-mvp-design.md`.

This service implements the minimum runnable version from that draft:

- text document ingest
- simple chunk splitting
- Ollama embeddings through `qwen3-embedding:0.6b`
- PostgreSQL + pgvector storage
- top-k source retrieval
- embedded admin console
- Cobra/Viper `server.yaml` runtime skeleton aligned with `gofurry-admin`
- JWT Cookie login for admin APIs
- overview console with database and Ollama health information

Deferred capabilities:

- file upload
- manual chunk editing
- reindex API
- natural-language answer generation
- rerank
- PDF, DOCX, OCR, and crawler ingestion
