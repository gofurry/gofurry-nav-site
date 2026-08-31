# Engineering documentation

This directory contains cross-service documentation for the active GoFurry production stack.

- [Repository architecture](architecture.md)
- [Local development](development.md)
- [Cross-service deployment](deployment.md)
- [Historical Fact operations](historical-facts.md)
- [Analytics Metric operations](analytics-metrics.md)
- [Change Intelligence operations](change-intelligence.md)
- [Public Insights contract](public-insights.md)
- [Admin identity and authorization](admin-identity.md)
- [React Admin development](admin-react.md)
- [Admin frontend parity](admin-frontend-parity.md)
- [Admin Data and System Operations](admin-data-system-operations.md)
- [Admin role operator guide](operations/admin-roles.md)
- [Collection Center operations](operations/collection-center.md)
- [V3 P0 production acceptance](acceptance/v3-p0-production-acceptance.md)
- [Linux/systemd operations](operations/systemd.md)
- [Architecture decisions](decisions/README.md)
- [Compatibility contract](../contracts/compatibility.md)
- [Database contract](../contracts/database.md)
- [Availability contract](../contracts/availability.md)
- [Upstream dependency contract](../contracts/upstream.md)
- [Admin authorization contract](../contracts/authorization.md)
- [Admin frontend contract](../contracts/admin-frontend.md)

Application-specific documents remain beside their owners under `apps/cn`. Operational files that are consumed directly by Nginx or Coraza live under `ops`. Archive documentation under `legacy` is historical and is not an active operator guide.
