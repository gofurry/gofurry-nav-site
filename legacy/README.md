# Legacy modules — archive only

This directory contains decommissioned historical modules:

- `gofurry-rag`: former standalone RAG service and console;
- `gofurry-nav-frontend-legacy`: former Vue public frontend;
- `gofurry-ops-agent` and `gofurry-ops-center`: former operations services.

No active production application, build target, CI job, vulnerability scan, deployment tool, documentation example, or dependency may depend on `legacy/**`.

The repository does not modernize, test, package, or deploy these modules by default. Do not update their Go versions or dependencies as part of active-stack maintenance. Inspection is allowed, but revival requires a separate migration that establishes current ownership, security review, tests, and explicit production wiring.
