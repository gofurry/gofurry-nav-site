# Local development

Each application is an independent module or frontend project. Run Go commands from the corresponding directory under `apps/cn`, and run repository database/tooling checks from `tools`.

The five active applications and `tools` require Go 1.26.7.

The root repository intentionally has no `go.work`; dependencies and tests stay scoped to their owning modules.
