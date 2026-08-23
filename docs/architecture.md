# Repository architecture

Production code is organized under `apps/cn`: five independently built Go applications and the Nuxt frontend. `apps/intl` is reserved for the international site and is not an active production target.

Database ownership remains at the repository root under `db/`: Goose owns schema migrations, while each active Go module contains its sqlc queries and generated code. `ops/` contains deployment assets; `legacy/`, `experimental/`, and `third-party/` are outside the active production build.
