# Deployment overview

Build and deploy each active application independently. The root `build.bat` builds the five Linux Go binaries; the Nuxt frontend keeps its separate deployment flow under `apps/cn/nav-web`.

Schema changes are deployed explicitly with Goose from `db/`. Applications never run migrations during startup. Production rollout details are completed in the operator documentation as part of this cleanup.
