## Summary

- 

## Scope

- [ ] `gofurry-nav-backend`
- [ ] `gofurry-nav-web`
- [ ] `gofurry-admin`
- [ ] `gofurry-rag`
- [ ] Collector / scheduler / sync
- [ ] SQL / migration
- [ ] Documentation
- [ ] CI / deployment

## Compatibility

- [ ] No public API change
- [ ] Compatible public API addition
- [ ] v1 to v2 migration or deprecation
- [ ] Behavior change documented below
- [ ] Potential breaking change documented below

Notes:

## Validation

- [ ] `cd gofurry-nav-backend && go test ./...`
- [ ] `cd gofurry-nav-web && npm run typecheck`
- [ ] `cd gofurry-nav-web && npm run build`
- [ ] `cd gofurry-admin && go test ./...`
- [ ] `cd gofurry-admin/web && npm run build`
- [ ] `cd gofurry-rag && go test ./...`
- [ ] `cd gofurry-rag/web && npm run build`
- [ ] Not run, explained below

Notes:

## Data and Rollout

- [ ] No schema or data change
- [ ] SQL or migration included
- [ ] Config change required
- [ ] Deployment order matters
- [ ] Background jobs / sync behavior affected

Notes:

## Security and Secrets

- [ ] No real API keys, access tokens, refresh tokens, cookies, proxy passwords, DSNs, JWTs, or credential-bearing URLs are included.
- [ ] New logs, examples, diagnostics, and screenshots avoid exposing secrets or private data.
- [ ] External requests, search suggestions, sync jobs, or crawlers were reviewed for abuse and rate-limit risk where relevant.

## Documentation

- [ ] Docs updated
- [ ] Roadmap / assessment / migration notes updated
- [ ] Admin or operator workflow documented
- [ ] Documentation not needed
