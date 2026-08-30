# Admin frontend contract

`apps/cn/admin/react` is the target Admin frontend. Its frozen foundation is React 19, Vite, strict TypeScript, React Router, Tailwind CSS v4, shadcn-style wrappers over Base UI primitives, TanStack Query/Table, React Hook Form, Zod, ECharts, and Lucide.

## Architecture boundaries

UI layers flow in one direction:

~~~text
components/ui -> components/admin -> features
~~~

TanStack Query owns server state. Route filters, search, pagination, sorting, and workspace tabs use URL state where practical. React local state owns transient UI. Do not introduce a second server-state copy or a role-based client store.

Themes are `system`, `light`, and `dark`, defaulting to `system`. Business components consume semantic tokens: background, surface, surface-muted, foreground, muted-foreground, border, primary, success, warning, danger, and info.

The backend authorization contract remains authoritative. Navigation and actions ask whether the current principal has a capability; frontend code must not reproduce the Role-to-Capability mapping.

## Product structure

Top-level groups are Workbench, Nav Content, Game Content, Data Operations, and System. P0.5.2-B makes content routes native React:

~~~text
/nav/sites
/nav/sites/:id
/nav/site-groups
/nav/update-notices
/nav/sayings
/game/games
/game/games/:id
/game/tags
/game/comments
/game/prizes
~~~

Site and Game are dedicated workspaces. Simple resources use the typed Resource Engine. Persistence mapping tables are managed as relationships inside workspaces, not exposed as primary navigation.

Collection, Metric, Change, Data Operations, Audit, and Accounts remain compatibility links or explicit placeholders until their owning stage. The Vue frontend under `apps/cn/admin/web` remains the embedded production frontend until P0.5.2-D; no React change may silently switch production routing.
