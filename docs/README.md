# Engineering documentation

This directory contains cross-service documentation for the active GoFurry production stack.

- [Repository architecture](architecture.md)
- [Local development](development.md)
- [Cross-service deployment](deployment.md)
- [Linux/systemd operations](operations/systemd.md)
- [Database migration contract](../contracts/database-migrations.md)
- [Production rollout contract](../contracts/production-rollout.md)

Application-specific documents remain beside their owners under `apps/cn`. Operational files that are consumed directly by Nginx or Coraza live under `ops`. Archive documentation under `legacy` is historical and is not an active operator guide.
