# GoFurry navigation repository

Production code is limited to five Go modules plus one frontend:

- `gofurry-game-collector` and `gofurry-game-backend` own Game runtime behavior.
- `gofurry-nav-collector` and `gofurry-nav-backend` own Nav runtime behavior.
- `gofurry-admin` uses explicit `gfa`, `gfn`, and `gfg` PostgreSQL pools.
- `gofurry-nav-web` is the production Nuxt frontend.
- `db/game`, `db/nav`, and `db/admin` own Goose migrations for `gfg`, `gfn`, and `gfa`.

Read `.agents/architecture.md`, `.agents/playbook.md`, and `contracts/` before database work. Use `rg` to locate any deeper module documentation.

Hard rules: Goose is the only schema owner; sqlc is the normal SQL contract; production PostgreSQL uses pgx/v5 + pgxpool; generated sqlc code is committed and never hand-edited. Do not add an ORM, generic repository/UnitOfWork/query builder, startup migrations, Redis key changes, collector scheduling changes, forced Nav V1 removal, or modernization of `legacy`, `experimental`, or `third-party`.
