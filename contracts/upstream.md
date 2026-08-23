# Upstream dependency contract

- Steam integrations validate identifiers and response shapes, use explicit timeouts, and bound response sizes. Retry is limited to safe operations with bounded backoff; rate limits and `Retry-After` are respected.
- External HTTP, DNS, and TLS failures remain attributable to the upstream boundary. Logs and returned errors preserve useful provenance without exposing credentials or private request data.
- GeoIP data is a versioned operational input. Missing, unreadable, or stale data is reported explicitly rather than replaced with fabricated location data.
- Parsers accept only known compatible variations, fail safely on malformed input, and use fixtures to validate upstream shape changes.
- `third-party/steam-go` is reference or maintained source outside the active module graph. Active applications consume a published dependency; parser changes require compatibility validation before publication and adoption.
- Concurrency must not bypass upstream rate limits. Retries, caching, and fallback behavior must preserve public API and collector contracts.
