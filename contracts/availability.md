# Availability contract

- `GET /livez` reports only whether the process is alive and able to serve the probe. It does not call external or business dependencies.
- `GET /readyz` reports whether the process is ready for its local runtime role. Initialization, shutdown, or failure of a locally required PostgreSQL, Redis, scheduler, or Bbolt dependency may return `503`.
- Admin `GET /startupz` reports completion of Admin bootstrap separately from liveness and readiness.
- Collector health listeners are internal operational interfaces. When enabled, they bind to loopback or a private address and are not exposed publicly by default.
- Failure of an external collection target does not make a locally healthy Collector unready. Target failures remain visible through collection results and logs.
- `gf-uptime` is an independent observer backed by a durable local Bbolt file, with no external/business Redis runtime dependency. Its readiness covers its own HTTP and Bbolt runtime, not the health of monitored targets.
- Public external monitoring observes `gf-uptime`; `gf-uptime` observes the configured application endpoints. Monitoring must not create a dependency from business services back to the observer.
