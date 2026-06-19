# Legacy Modules

This directory contains modules that have been removed from the active gofurry runtime path.

## Archived Modules

- `gofurry-rag`: Former standalone RAG service and console. The public Q&A page, Nuxt proxy routes, and default build/deploy wiring have been removed from the active stack.
- `gofurry-nav-frontend-legacy`: Former Vue frontend kept as a migration reference.

## Maintenance Rules

- Do not include these modules in the root `build.bat all` target or the default CI matrix.
- Do not add new production dependencies from active services to this directory.
- If a legacy module must be inspected or revived, treat it as a separate migration task and document the new ownership before wiring it back into active services.
