# Admin CloudOps: daily EdgeOne main-host purge

Issue #122 adds one operational YAML schedule to the existing Admin runtime.
There is no UI, new service, dependency, or database migration. Rebuild/redeploy
Admin, then restart it with its explicit `serve --config <file>` configuration.
Existing files without `scheduled_purge` remain disabled.

## Configuration

Add this inside the existing EdgeOne configuration, preserving its credentials,
zone and hosts. Production operators may enable the following schedule:

```yaml
external_services:
  cloud_ops:
    edgeone:
      scheduled_purge:
        enabled: true
        time: "05:30"
        timezone: "Asia/Shanghai"
```

The shipped example uses `enabled: false`. `time` must be exactly `HH:mm`;
`timezone` must be an explicit IANA name (host-dependent `Local` is rejected).
The binary bundles timezone data. When enabled, missing EdgeOne credentials,
disabled provider, missing/invalid main host or invalid schedule fails startup.
Startup logs host, time, timezone and next run, never credentials.

Only the existing `edgeone.main_host` is submitted through the ordinary,
allowlisted `host` purge. Asset host, Cloudflare and whole-zone purge are not
scheduled. Manual Cloud Resources controls and capabilities are unchanged.

## Execution and failure handling

- No startup execution or catch-up: starting at 06:00 after a missed 05:30
  schedules tomorrow's 05:30.
- One submission attempt per daily slot; no automatic retry, even on timeout.
  The provider retains its 15-second timeout and disabled SDK retries.
- Scheduled work has a 25-second context. Shutdown cancels it, allows up to
  3 seconds for outcome audit and 2 seconds for unlock, then closes Redis/DB and
  flushes logs. `Stop` is idempotent.
- A dedicated session advisory lock (`718210552122`) uses GFA and keeps its
  owning connection through audit and submission. A busy lock causes a logged
  skip without duplicate operation audit. An uncertain unlock closes that
  connection instead of returning it to the pool.
- Under the lock, an existing same-host, same-slot intent suppresses later
  instances, including those arriving after the first finishes. All replicas
  must share GFA and the same host/time/timezone configuration and use session
  connections (not a transaction-pooling proxy). Keep clocks synchronized.
  Retain same-slot audit records; deleting them removes the durable duplicate
  guard. No new table or schema permission is needed by `gofurry_app`.

Audit action is `cloud.edgeone.purge.host.scheduled`, resource `cloud`, target
the main host, operator `system`. Both records share a request ID such as
`edgeone-scheduled-purge:20260919T0530+0800`:

1. `requested` includes `type=host`, the one-element `targets`, and the scheduled
   RFC3339 timestamp with timezone offset. This is committed before submission.
   If it fails, nothing is submitted.
2. `completed` includes the returned job ID and `status=submitted`, or `failed`
   records the sanitized error. `completed` means submission was accepted, not
   that the CDN finished purging; inspect the existing EdgeOne task history.

A result-audit failure is logged and leaves the intent as the durable guard;
never automatically replay it. A crash after intent may leave only `requested`.
Inspect task history before deciding whether a separate manual operation is
needed. Disable the YAML schedule and restart Admin to stop future submissions.

## Development validation

Run from the developer workstation, using development infrastructure only.
Never enable this against a configuration whose `main_host` is production.
The existing development EdgeOne zone also contains production hosts; a
development asset bucket/hostname alone does **not** make main-host purge safe.

Unit tests and PostgreSQL acceptance from `apps/cn/admin`:

```powershell
go test ./config ./internal/app/cloudops ./internal/infra/cloudops ./internal/bootstrap -count=1
$env:GOFURRY_ADMIN_SCHEDULER_TEST_CONFIG = (Resolve-Path config/server.yaml).Path
go test ./internal/app/cloudops -run '^TestScheduledPurgePostgres$' -count=1 -v
```

The opt-in test requires the development GFA `gofurry_app` account. It proves
A-lock/B-skip/A-unlock/B-lock across separate one-connection pools, checks audit
and slot lookup on the held connection in a rolled-back transaction, and checks
session disposal after failed unlock. No DDL, persistent audit rows, or cloud
calls occur. A skipped test is not acceptance. The test-only environment variable
selects an explicit YAML; it is not a new runtime environment setting.

For live end-to-end acceptance, an operator must first supply an approved,
EdgeOne-served **development main hostname**, plus a matching ignored local YAML:

1. Confirm `main_host` is that development hostname. Set an enabled schedule a
   few minutes ahead in its explicit timezone. Run Admin locally with that YAML.
2. Verify startup reports the correct target/next run and submits nothing yet.
3. At the slot, verify exactly one host-purge job in Cloud Resources task history
   and the two matching system audit records in GFA Audit.
4. On a subsequent near-future slot, run two local Admin instances using distinct
   HTTP ports, the same GFA and identical EdgeOne schedule/host. Verify one remote
   job and one audit pair; the other instance logs a lock/recorded-slot skip.
5. Stop both processes and restore disabled scheduling in the local YAML.

The existing `TestRealDevCloudOps` separately covers manual development asset
purges and provider history; it intentionally never purges the production main
host. It is not a substitute for the live daily-schedule acceptance above.
