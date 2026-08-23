# Engineering documentation

This directory contains cross-service documentation for the active GoFurry production stack.

- [Repository architecture](architecture.md)
- [Local development](development.md)
- [Cross-service deployment](deployment.md)
- [Linux/systemd operations](operations/systemd.md)
- [Architecture decisions](decisions/README.md)
- [Compatibility contract](../contracts/compatibility.md)
- [Database contract](../contracts/database.md)
- [Availability contract](../contracts/availability.md)
- [Upstream dependency contract](../contracts/upstream.md)

Application-specific documents remain beside their owners under `apps/cn`. Operational files that are consumed directly by Nginx or Coraza live under `ops`. Archive documentation under `legacy` is historical and is not an active operator guide.
