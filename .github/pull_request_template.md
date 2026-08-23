## Summary

-

## Scope

- [ ] Nav Web
- [ ] Nav Backend
- [ ] Nav Collector
- [ ] Game Backend
- [ ] Game Collector
- [ ] Admin
- [ ] Uptime / Availability
- [ ] Database / Migration
- [ ] CI / Tooling / Ops
- [ ] Documentation
- [ ] Cross-service

## Compatibility

- [ ] Public API, Redis, database ownership, and Collector behavior are unchanged
- [ ] Compatible behavior is added and documented
- [ ] An intentional behavior change is explained below

Notes:

## Validation

- [ ] Affected Go tests passed
- [ ] Frontend typecheck/build passed when applicable
- [ ] sqlc/Goose checks passed when applicable
- [ ] Integration tests passed when applicable
- [ ] Repository policy passed
- [ ] Not fully verified — explained below

Notes:

## Data and Operations

- [ ] No schema or data change
- [ ] Schema/data changes use Goose and include recovery evidence
- [ ] No configuration or deployment change
- [ ] Configuration or deployment impact is explained below

Notes:

## Security and Secrets

- [ ] No credentials, private data, or credential-bearing URLs are included
- [ ] External inputs, logs, and rate limits were reviewed when applicable

## Documentation

- [ ] Contracts or accepted ADRs were updated when a durable boundary changed
- [ ] User, developer, or operator documentation was updated when applicable
- [ ] Documentation is not needed
