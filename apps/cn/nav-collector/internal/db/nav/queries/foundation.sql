-- name: FoundationPing :one
SELECT 1::bigint AS value;

-- name: ListCollectorDomains :many
SELECT cd.id, COALESCE(cd.site_id, 0)::bigint AS site_id, cd.name, cd.proxy,
       cd.prefix, cd.tls, cd.deleted
FROM gfn_collector_domain cd
JOIN gfn_site s ON s.id = cd.site_id
WHERE cd.deleted IS NOT TRUE AND cd.site_id > 0 AND s.deleted IS NOT TRUE
ORDER BY cd.site_id, cd.id;

-- name: InsertPingLog :exec
INSERT INTO gfn_collector_log_ping (id, name, delay, loss, status, create_time)
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.arg(delay), sqlc.arg(loss),
        sqlc.arg(status), sqlc.arg(create_time));

-- name: InsertHTTPLog :exec
INSERT INTO gfn_collector_log_http (id, name, info, status, create_time)
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.arg(info)::jsonb,
        sqlc.arg(status), sqlc.arg(create_time));

-- name: InsertDNSLog :exec
INSERT INTO gfn_collector_log_dns (
    id, name, a, aaaa, mx, ns, soa, txt, caa, cname, status, create_time
) VALUES (
    sqlc.arg(id), sqlc.arg(name), sqlc.narg(a)::jsonb, sqlc.narg(aaaa)::jsonb,
    sqlc.narg(mx)::jsonb, sqlc.narg(ns)::jsonb, sqlc.narg(soa)::jsonb,
    sqlc.narg(txt)::jsonb, sqlc.narg(caa)::jsonb, sqlc.narg(cname)::jsonb,
    sqlc.arg(status), sqlc.arg(create_time)
);

-- name: InsertObservation :exec
INSERT INTO gfn_collector_observation (
    id, site_id, target, protocol, status, observed_at, duration_ms,
    error_code, error_message, payload, schema_version, create_time,
    job_id, run_id, collector_instance_id
) VALUES (
    sqlc.arg(id), sqlc.arg(site_id), sqlc.arg(target), sqlc.arg(protocol),
    sqlc.arg(status), sqlc.arg(observed_at), sqlc.narg(duration_ms),
    sqlc.narg(error_code), sqlc.narg(error_message), sqlc.arg(payload)::jsonb,
    sqlc.arg(schema_version), sqlc.arg(create_time), sqlc.narg(job_id),
    sqlc.narg(run_id), sqlc.narg(collector_instance_id)
);

-- name: ListObservationTrendRows :many
SELECT protocol, status, observed_at, COALESCE(duration_ms, 0)::bigint AS duration_ms,
       error_code, payload::text AS payload
FROM gfn_collector_observation
WHERE site_id = sqlc.arg(site_id) AND target = sqlc.arg(target)
  AND observed_at >= sqlc.arg(since)
  AND protocol IN ('ping', 'http', 'dns')
ORDER BY observed_at DESC, id DESC
LIMIT sqlc.arg(limit_count);

-- name: ListObservationChangeRows :many
SELECT protocol, status, observed_at, COALESCE(duration_ms, 0)::bigint AS duration_ms,
       error_code, payload::text AS payload
FROM gfn_collector_observation
WHERE site_id = sqlc.arg(site_id) AND target = sqlc.arg(target)
  AND observed_at >= sqlc.arg(since)
  AND protocol IN ('http', 'dns', 'port_check', 'rdap')
ORDER BY observed_at DESC, id DESC
LIMIT sqlc.arg(limit_count);
