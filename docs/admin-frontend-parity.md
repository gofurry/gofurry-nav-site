# Admin frontend functional parity

This matrix was produced from the actual `dev` implementation at the start of P0.5.2-D, before deleting `apps/cn/admin/web`. Parity means that the still-valid workflow is available, not that the Vue page layout is reproduced.

| Legacy workflow | React disposition | Evidence / cutover note |
|---|---|---|
| Bootstrap, login, logout, current identity | React equivalent exists | `/setup`, `/login`, the App Shell account menu, and the existing auth/CSRF transport use the same backend contract. |
| Legacy landing page | React intentionally supersedes legacy UX | `/` is the capability-shaped Workbench instead of redirecting to Sayings. |
| Sayings and update notices CRUD | React equivalent exists | Schema-driven Resource Engine definitions. |
| Sites CRUD | React intentionally supersedes legacy UX | Site list plus Site Workspace content tabs. |
| Collector-domain CRUD and Primary selection | React intentionally supersedes legacy UX | Site Workspace Collection tab owns target create/edit/delete, Primary selection, and immediate collection. |
| Site-group mapping bulk replace | React intentionally supersedes legacy UX | Site Workspace Classification tab edits the complete group relationship set with the existing bulk-replace API. |
| Featured-site CRUD | React intentionally supersedes legacy UX | Site Workspace Classification tab owns featured state and weight. |
| Site groups CRUD | React equivalent exists | Schema-driven Resource Engine definition. |
| Games CRUD and Steam prefill | React equivalent exists | Game Workspace Content tab preserves Steam prefill and the existing game payload. |
| Game-tag mapping bulk replace by game | React intentionally supersedes legacy UX | Game Workspace Classification tab edits the complete relationship set with the existing bulk-replace API. |
| Reverse tag-map page / bulk replace by tag | React intentionally supersedes legacy UX | Persistence-shaped reverse navigation is removed by contract; Game Workspace is the canonical relationship editor and Tag remains a simple resource. |
| Tags, comments, and prizes CRUD | React equivalent exists | Schema-driven Resource Engine definitions. |
| Remote option lookup/search | Missing at audit; resolved in D | Resource Engine remote fields now query the existing paged option endpoints by search text instead of loading only a fixed first page. |
| Resource list search, pagination, create/edit/delete | React equivalent exists | Resource Engine uses backend pagination, validation, server-confirmed mutations, and destructive confirmation. |
| Collection overview and Collector lifecycle | React equivalent exists | Overview, Current/Historical instance views, attention state, and charts are native. |
| Collection schedules, Run Now, edit/toggle | React equivalent exists | Native capability-aware controls preserve schedule and Run Now semantics. |
| Running queue and cancel | React equivalent exists | Native queue view; cancel requires `collection.control` and explicit impact confirmation. |
| Run history, chart filters, task-result filters | Missing at audit; resolved in D | React restores window/job/domain/time filters and domain-specific result filters with count-backed pagination. |
| Manual collection and retry | React equivalent exists | Searchable Game/Site/Target selection and the existing job/retry endpoints. |
| Metric overview, daily results, entity states | React equivalent exists | Native business views preserve unknown/null semantics and real pagination. |
| Metric Registry and checkpoints | React equivalent exists | Read-only technical tab guarded by `metrics.technical`. |
| Change events and entity history | React equivalent exists | Native recent/event and entity-oriented views with detail provenance. |
| Change Registry and checkpoints | React equivalent exists | Read-only technical tab guarded by `changes.technical`. |
| Data Operations | Not applicable in legacy; React is canonical | Read-only `dataops.read` view for bounded gfa/gfn/gfg metadata. |
| Audit | Not applicable in legacy; React is canonical | Snapshot-aware, redacted, count-backed audit history. |
| Accounts & Permissions | Not applicable in legacy; React is canonical | Owner-only account governance with backend last-active-Owner protection. |
| Theme and global search | Not applicable in legacy; React is canonical | System/Light/Dark and keyboard-accessible global search. |

No parity repair changes Fact, Metric, Change, Collection, or RBAC semantics. After the two audit gaps above were resolved and regression-tested, no still-valid workflow depended on Vue.
