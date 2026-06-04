# gofurry-rag Production Readiness Code Audit

Date: 2026-05-17

## Summary

`gofurry-rag` has completed the planned v1.0 feature scope and the production-readiness findings from the previous audit have been fixed in this pass.

Current status:

- P0 Critical: 0 open
- P1 High: 0 open
- P2 Medium: 0 open
- P3 Low: 0 open

## Completion Check

The roadmap is functionally complete through v1.0:

- v0.6 public chat boundaries: implemented.
- v0.7 frontend citation semantics: implemented.
- v0.8 site content sync: implemented.
- v0.9 nav/games scenario entry points: implemented.
- v1.0 lightweight multi-turn context: implemented.

The prompt now treats recent conversation history only as reference context for follow-up understanding. Facts must come from the current retrieval results.

## Fixed Findings

### P1-001: Unconditional proxy trust can weaken public IP rate limiting

Status: Fixed.

Changes:

- Added explicit proxy trust configuration.
- Default behavior no longer trusts proxy headers.
- Fiber `TrustProxyConfig` is now populated from configured trusted proxy addresses or trusted ranges.
- Public rate limiting continues to use `c.IP()`, but `c.IP()` is now protected by Fiber proxy trust rules.

Production note:

If the service runs behind Nginx, Caddy, or another reverse proxy, set `server.trust_proxy=true` and configure `server.trusted_proxies` or the appropriate `server.trust_proxy_*` range. Otherwise all visitors may be grouped under the reverse proxy IP.

### P2-001: Public chat limiter keeps one map entry per observed IP without cleanup

Status: Fixed.

Changes:

- Expired limiter windows are cleaned before accepting a new key.
- The public limiter now has a bounded key count.
- When the limiter is full and no expired keys can be removed, new keys are rejected instead of growing memory indefinitely.

### P2-003: Public query validation can do database work before applying the public rate limiter

Status: Fixed.

Changes:

- Public query limits are now built from loaded config/defaults.
- `NewServer` and public request validation no longer call `ChatStatus()` to fetch limits.
- Public rate limiting runs before public request length and parameter validation.
- Public query validation no longer triggers `repo.Overview()`.

### P2-004: Embedding dimension is configurable but the schema is hard-coded to `vector(1024)`

Status: Fixed.

Changes:

- Config validation now rejects `rag.embed_dim` values other than `1024`.
- The current database schema remains explicitly tied to `rag_chunks.embedding vector(1024)`.

Production note:

Changing embedding dimension still requires a deliberate database migration. It is no longer possible to accidentally start the service with an incompatible dimension.

### P3-001: SSE errors try to set HTTP status after streaming has begun

Status: Fixed.

Changes:

- `chatStream` no longer attempts to change HTTP status after `SendStreamWriter` starts.
- Stream-time failures are reported through `event: error`.
- Error payloads include both `status` and `message`.
- Pre-stream failures still use normal HTTP status codes.

### P3-002: Error response bodies from upstream model/sync services are read without a cap

Status: Fixed.

Changes:

- nav sync upstream error bodies are capped at 64 KiB.
- game sync upstream error bodies are capped at 64 KiB.
- Tencent model upstream error bodies are capped at 64 KiB.

## Verification

Validated with:

```powershell
cd gofurry-rag
go test ./...
go vet ./...
```

Both commands pass.

## Remaining Production Checklist

- Use production-only secrets and do not deploy the committed local `server.yaml` as-is.
- Enable secure cookies outside debug mode.
- Configure trusted proxy settings only for the actual reverse proxy addresses.
- Keep database backups before the production update.
