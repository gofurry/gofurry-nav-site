# Game V2 read-model persistence

Prize, review, and view-count SQL is sqlc-managed. The V2 aggregate read model
is a documented direct-pgx exception because it combines a fixed projection
with bounded dynamic search filters, sort variants, language fallback, and
ordered batch hydration into the existing public response model. Expressing
that composition in sqlc would multiply near-identical queries and mapping
types without changing the database contract.

The exception is limited to this package. SQL fragments are fixed in source,
all request values use PostgreSQL parameters, sort expressions come from
closed internal variants, and PostgreSQL characterization tests cover the
public detail/list/search/review/tag/recommendation/collector-status behavior.
This package is not a generic repository or a SQL builder.
