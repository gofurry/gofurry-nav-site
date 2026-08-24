# 0005: Normalize the Game release and language domain

## Status

Accepted

## Context

Game V2 preserved Steam release and language fields as localized strings. Those raw values are useful for compatibility and debugging but cannot safely define release ordering, date-range search, a game's first playable date, or normalized language capabilities.

## Decision

Use the unversioned `gfg_game_release_state`, `gfg_game_first_available`, `gfg_game_release_history`, and `gfg_game_languages` tables as the canonical Game domain. Goose owns their schema and sqlc owns normal static SQL.

The US/English Storefront response is the only automatic canonical source. Collector payload locale is represented by `StoreLocale`; it is not a game language. Current Release State is a replaceable observation, while First Available means the first date or date range when a game was formally purchasable or playable, including Early Access. Automatic First Available writes are write-once. Release History records normalized semantic changes, not raw-text-only changes.

The existing V2 pipeline and `/api/v2/game/*` routes evolve in place. Raw `release_date_text` and `supported_languages` remain compatibility/upstream fields, and `gfg_game.release_date` remains temporarily for audit and rollback, but business queries and new UI behavior do not depend on them.

## Alternatives Considered

Parsing localized strings in each reader, versioning canonical table names, copying the pipeline into V3 packages, treating current release state as first availability, and implementing a general change-intelligence framework were rejected because they create ambiguous facts, duplicated runtime paths, or P0.2+ scope.

## Consequences

Latest Games and release-range search use First Available windows. Detail responses expose typed calendar dates and languages while retaining compatibility fields. AppID correction transactionally clears Steam-derived state and re-enqueues the existing V2 collector. Later analytics can consume canonical facts without guessing legacy strings.
