# GoFurry Development Infrastructure

> Scope: shared development infrastructure for the GoFurry monorepo.
>
> This document is intentionally safe for the public repository. It describes architecture, access conventions, operating boundaries, and recovery expectations, but does **not** contain passwords, tokens, private keys, public IPs, Tailnet addresses, or other live credentials.

## 1. Purpose

The development server is **infrastructure only**. It is not a shared remote workstation and it is not a production host.

It provides shared stateful dependencies and supporting services for local development:

- PostgreSQL
- Redis
- MongoDB for other/future GoFurry development needs
- Tailscale private access
- Docker / Docker Compose
- Mihomo outbound proxy for dependency and image access from the mainland China VPS
- Nginx installed for possible future internal reverse-proxy use

Developers normally run Go, Nuxt, React, collectors, tests, and other application code on their own machines.

Ordinary developers do **not** receive:

- SSH access
- sudo access
- Docker access on the server

Infrastructure administration is restricted to the project owner and explicitly authorized operations staff.

## 2. Source of truth

Repository-level architecture and development contracts remain authoritative:

- `AGENTS.md`
- `.agents/architecture.md`
- `.agents/playbook.md`
- `docs/development.md`
- `contracts/database.md`
- `db/README.md`
- this document

The server also keeps a private, instance-specific operations manual:

```text
/srv/gofurry-dev/docs/gofurry-dev-infra-operations-manual-2026-09-06.md
/srv/gofurry-dev/docs/developer-access.md
```

These server-side documents may contain more operational detail than the public repository document.

When Codex or another authorized agent performs server maintenance, it should read this repository document first, then read the server-side manual after SSH login before making substantial changes.

## 3. Operator access

Use a stable local SSH alias rather than committing a Tailnet IP or MagicDNS hostname to the repository.

Recommended local SSH configuration:

```sshconfig
Host gofurry-dev-infra
    HostName <TAILNET_MAGICDNS_OR_IP>
    User ops
    IdentityFile ~/.ssh/<PRIVATE_KEY>
    IdentitiesOnly yes
```

Normal operator login:

```bash
ssh gofurry-dev-infra
```

The repository must never contain the real private key, Tailscale auth key, database password, API token, or other secret required for access.

If the SSH alias is not available on the operator machine, stop and ask the operator for the correct local access configuration instead of guessing or falling back to a public address.

## 4. Server baseline

Current baseline as of 2026-09-06. Treat versions as a snapshot and verify them before depending on an exact value.

| Component | Baseline |
|---|---|
| OS | Debian 13.6 |
| Hostname | `gofurry-dev-infra` |
| Kernel | `6.12.107+deb13-amd64` |
| CPU / RAM | 2 vCPU / 4 GiB |
| Root disk | about 100 GiB |
| Swap | 2 GiB emergency swap, low swappiness |
| Docker Engine | 29.8.0 |
| Docker Compose | 5.5.1 |
| Go | 1.27.1 |
| Node.js | 24.20.0 LTS via nvm |
| Python | Debian system Python remains untouched; CPython 3.14.7 is managed by uv |
| PostgreSQL client | 18.6 |
| Redis client | 8.0.2 |
| Nginx | 1.31.5 mainline |
| Mihomo | 1.19.30 |

The infrastructure root is:

```text
/srv/gofurry-dev
```

Expected high-level layout:

```text
/srv/gofurry-dev/
├── .env
├── compose.yml
├── backups/
├── configs/
├── credentials/
├── docs/
└── scripts/
```

`.env` and `credentials/` are private operational state and must never be committed or printed into logs, issues, pull requests, or chat transcripts.

## 5. Network model

The database services are not intended to be publicly reachable.

The expected path is:

```text
developer workstation
    ↓
Tailscale encrypted network
    ↓
tailscaled TCP forwarding
    ↓
127.0.0.1 on the VPS
    ↓
Docker-published localhost port
    ↓
database container
```

Current service ports:

| Service | Server-local binding |
|---|---:|
| PostgreSQL | `127.0.0.1:5432` |
| Redis | `127.0.0.1:6379` |
| MongoDB | `127.0.0.1:27017` |

Tailscale exposes the corresponding TCP services only to the Tailnet.

Verify rather than assume:

```bash
tailscale status
tailscale serve status
ss -lntp
```

### Network safety rules

Do not:

- publish database ports on `0.0.0.0`
- open PostgreSQL, Redis, or MongoDB in the Tencent Cloud security group for normal developer access
- replace Tailscale with direct public database exposure
- enable Tailscale Exit Node or Tailscale SSH without an explicit project decision
- assume a Tailnet IP or MagicDNS name from documentation; discover the current live value on the authorized machine/server

Application-level database TLS is not currently required for the private Tailnet path; Tailscale provides transport encryption. Revisit this only if the network boundary changes.

## 6. Docker services

The Compose project lives at:

```bash
cd /srv/gofurry-dev
```

Current core images:

```text
pgvector/pgvector:0.8.6-pg18-trixie
redis:8.10.1-trixie
mongo:8.3.8-noble
```

Basic inspection:

```bash
cd /srv/gofurry-dev
docker compose ps
docker compose images
docker stats --no-stream
docker system df
```

Validate Compose without printing rendered configuration:

```bash
docker compose config >/dev/null
```

Inspect one service:

```bash
docker compose logs --tail=200 postgres
docker compose logs --tail=200 redis
docker compose logs --tail=200 mongodb
```

Recreate only the service being changed when possible:

```bash
docker compose up -d --no-deps --force-recreate <service>
```

### Destructive Docker operations

Do **not** run any of the following unless the operator explicitly asks for data destruction and the impact has been reviewed:

```bash
docker compose down -v
docker volume rm ...
docker system prune -a
rm -rf /var/lib/docker
rm -rf /var/lib/containerd
```

Never destroy or recreate a database volume merely to troubleshoot configuration.

## 7. PostgreSQL model

The active monorepo owns three PostgreSQL databases:

```text
gfn  -> Nav
gfg  -> Game
gfa  -> Admin authentication and audit state
```

Goose under the repository root is the sole schema migration source of truth:

```text
db/nav/migrations
db/game/migrations
db/admin/migrations
```

Applications must not create or migrate schema at startup.

The shared cluster currently has all three databases, initialized from empty databases through the repository's Goose migrations. The current privilege model is:

```text
infrastructure admin
    └── owns infrastructure-level database/extension administration

gofurry_migrator
    └── Goose / schema migration authority

gofurry_app
    └── normal application DML

gofurry_readonly
    └── read-only inspection
```

All runtime `server.yaml` files use `gofurry_app`. Goose uses `gofurry_migrator` only, GUI/manual inspection uses `gofurry_readonly`, and `gofurry_admin` remains infrastructure-administration-only. Do not bypass this separation for debugging convenience.

Current development data is intentionally selective:

- `gfn` contains production-derived public/current content needed for Nav development.
- `gfg` contains production-derived catalog/current read-model data needed for Game development.
- `gfa` deliberately contains no migrated production identity or audit data.

This is not a contract to bulk-copy production databases. Any later refresh must remain scoped, sanitized where necessary, and explicitly authorized.

Production credentials must never be reused in development.

## 8. Redis model

Redis is shared development infrastructure, not production cache state.

Development Redis starts from independent cache state; production Redis cache is not copied into this host.

The access model separates:

- infrastructure/admin access
- runtime application access through the `gofurry_app` ACL user

Application credentials do not receive Redis administrative or dangerous command categories. A Redis GUI receiving an authorization error for `INFO` while normal application commands work is expected least-privilege behavior, not evidence that the application ACL should be broadened.

Before changing ACLs, inspect the current live Redis configuration and the Compose definition. Never widen permissions merely to make an administrative GUI command succeed.

## 9. MongoDB model

MongoDB is available as shared development infrastructure for future or adjacent GoFurry projects.

The current active `gofurry-nav-site` runtime topology does not depend on MongoDB.

Do not manufacture a MongoDB dependency for this monorepo merely because the service exists on the shared development VPS.

## 10. Outbound proxy

The mainland development VPS uses Mihomo for outbound access where direct GitHub / Docker Hub connectivity is unreliable.

Expected host-side components:

```text
/usr/local/bin/mihomo
/etc/mihomo/config.yaml
mihomo.service
127.0.0.1:7890
```

Inspect:

```bash
systemctl status mihomo --no-pager
journalctl -u mihomo -n 100 --no-pager
ss -lntp | grep 7890
```

The `ops` shell provides helpers:

```bash
proxy_on
proxy_off
```

Docker daemon proxy settings are managed separately from the interactive shell.

Do not introduce random registry mirrors as a default workaround. Prefer the existing Mihomo path and verify it first.

Do not print proxy subscription URLs, node credentials, or other provider secrets.

## 11. Host services

Useful inspection:

```bash
hostnamectl
uname -a
df -h
free -h
swapon --show
systemctl --failed
systemctl status docker --no-pager
systemctl status tailscaled --no-pager
systemctl status mihomo --no-pager
```

Chrony and unattended APT upgrades are part of the host baseline.

Nginx is installed for possible future reverse-proxy use. Its current service state must be checked before assuming that it is active or disabled:

```bash
nginx -v
systemctl status nginx --no-pager
```

Do not add public listeners or reverse-proxy routes merely because Nginx is installed.

## 12. Backup and restore status

Backup automation and restore drills must be treated as operational contracts, not as assumptions.

The repository documents the required coverage and safe procedure, but contains no evidence proving that the current server backup automation or a restore drill has passed. Verify live server records before reporting either as successful.

The final PostgreSQL backup design must cover all active application databases:

```text
gfn
gfg
gfa
```

A valid restore drill must restore into temporary test databases or temporary volumes and verify the restored data.

Never test restore procedures by overwriting the active development database or active Redis/MongoDB volume.

Before changing backup automation, inspect:

```text
/srv/gofurry-dev/scripts
/srv/gofurry-dev/backups
/etc/systemd/system
```

and read the server-side private operations manual.

## 13. Codex operating protocol

When an authorized Codex session is asked to maintain this server, follow this order.

1. Read repository contracts first:

```text
AGENTS.md
.agents/architecture.md
.agents/playbook.md
docs/development.md
contracts/database.md
db/README.md
docs/operations/dev-infrastructure.md
```

2. Perform a non-mutating local repository audit.

3. Connect using:

```bash
ssh gofurry-dev-infra
```

4. Read the private server manual:

```bash
sed -n '1,240p' /srv/gofurry-dev/docs/gofurry-dev-infra-operations-manual-2026-09-06.md
sed -n '1,240p' /srv/gofurry-dev/docs/developer-access.md
```

If the files are longer, continue reading them in additional bounded ranges.

5. Inspect actual server state before making changes:

```bash
hostnamectl
systemctl --failed
tailscale serve status
cd /srv/gofurry-dev && docker compose ps
df -h
free -h
```

6. Prefer the smallest targeted change.

7. Validate immediately after every mutation.

8. Report:
   - what was observed
   - what was changed
   - validation evidence
   - remaining risks or follow-up work

### Codex safety boundaries

Codex must not:

- operate on production unless the user explicitly changes the scope
- reuse production credentials
- print `.env`, credential files, private keys, tokens, or proxy subscriptions
- commit live infrastructure secrets to the repository
- delete Docker volumes as a troubleshooting shortcut
- run broad prune commands
- expose database ports publicly
- change Tencent Cloud security groups without explicit authorization
- reboot the server unless required and explicitly approved
- run Goose against production
- silently replace repository database contracts with ad-hoc DDL
- modify generated sqlc files by hand

When a requested action conflicts with repository contracts or the live server state, stop and surface the conflict instead of forcing the change.

## 14. Developer access boundary

Normal developers use local tools and connect to shared services through Tailscale.

Typical clients include:

```text
DataGrip / Navicat     -> PostgreSQL
RedisInsight           -> Redis
MongoDB Compass        -> MongoDB when needed
```

The repository should document logical database names and role purposes, but not live passwords or Tailnet addresses.

A developer who only needs PostgreSQL, Redis, or MongoDB access does not need SSH, sudo, or Docker privileges on the VPS.

## 15. Public-document security policy

This file may be committed to the public GoFurry monorepo.

Keep the following out of Git:

```text
real Tailnet IPs
real Tailnet MagicDNS hostnames when not necessary
Tencent Cloud public IPs
SSH private keys
Tailscale auth keys
database passwords
Redis passwords
MongoDB passwords
JWT secrets
API keys
proxy subscriptions
provider credentials
contents of /srv/gofurry-dev/.env
contents of /srv/gofurry-dev/credentials/
```

Operational structure, service names, local ports, container images, role names, filesystem layout, and recovery procedures are not secrets by themselves and may be documented publicly.

## 16. Current database state

As of 2026-09-06, the earlier generic-database transition is complete: `gfn`, `gfg`, and `gfa` exist and were initialized from empty databases by Goose. The selective `gfn` and `gfg` development datasets and intentionally development-only `gfa` identity/audit state are described in section 7. Redis remains an independent development cache as described in section 8.

Backup coverage and restore-drill success are separate operational checks. Do not infer them from database existence or migration success, and never claim PASS without live evidence.
