# Collection Center operations

Collection Center reads the durable Game/Nav PostgreSQL control plane. Collector instance rows are lifecycle history, not disposable presence rows. The default Current view shows the newest lifecycle record for each logical `collector_id`; Historical shows older records. Health remains derived from `stopped_at` and the existing heartbeat thresholds.

Schedule **Run Now** means a manual execution of that schedule capability. The Job has `trigger=manual`, preserves `schedule_id` and `schedule_version`, and has no `scheduled_for`. It must not change `last_materialized_for` or `next_scheduled_for`. Schedule Last/Coverage uses related Jobs through this lineage; zero expected targets yields null coverage and the UI renders `—`, while an actual zero-success non-empty run renders `0%`.

Run History and Task Results return count-backed pages and preserve existing Game/Nav filters. Manual collection uses searchable Game and Site options; Target scope first selects a Site and then one of that Site's current non-deleted targets. PostgreSQL still receives the same internal `game_id`, `site_id`, and canonical target values.

Operational smoke:

1. Run a Schedule with Run Now and confirm its Job contains the schedule ID/version while the schedule's next slot is unchanged.
2. After completion, confirm Schedule Last updates and Coverage is `—` when expected is zero.
3. Switch Collector Instances between Current and Historical; do not delete historical rows merely because they are stopped.
4. Page through Run History and Task Results and verify filters preserve the total.
5. Create manual Game, Nav Site, and Nav Target Jobs through the search selectors.
