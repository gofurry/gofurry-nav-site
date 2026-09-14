# Upstream dependency contract

- Steam integrations validate identifiers and response shapes, use explicit timeouts, and bound response sizes. Retry is limited to safe operations with bounded backoff; rate limits and `Retry-After` are respected.
- External HTTP, DNS, and TLS failures remain attributable to the upstream boundary. Logs and returned errors preserve useful provenance without exposing credentials or private request data.
- GeoIP data is a versioned operational input. Missing, unreadable, or stale data is reported explicitly rather than replaced with fabricated location data.
- Parsers accept only known compatible variations, fail safely on malformed input, and use fixtures to validate upstream shape changes.
- `third-party/steam-go` is reference or maintained source outside the active module graph. Active applications consume a published dependency; parser changes require compatibility validation before publication and adoption.
- Concurrency must not bypass upstream rate limits. Retries, caching, and fallback behavior must preserve public API and collector contracts.
- Direct YAML usage in active modules uses `go.yaml.in/yaml/v4`, pinned to `v4.0.0-rc.6` for #72. Its classic `Marshal`, `Unmarshal`, and `Decoder` APIs retain v3-compatible defaults; keep those configuration semantics and validate example configs on upgrades. Viper and other dependencies may still bring YAML v2/v3 transitively; do not force-replace their internal parsers.
