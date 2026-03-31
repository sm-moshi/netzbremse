-- Make metric columns nullable: NULL means "no measurement" (failed test),
-- a numeric value means "measured result" (including genuine zero).

-- Remove defaults that conflate "unknown" with "measured zero".
alter table measurements alter column download_bps drop default;
alter table measurements alter column upload_bps drop default;
alter table measurements alter column latency_ms drop default;
alter table measurements alter column jitter_ms drop default;
alter table measurements alter column download_latency_ms drop default;
alter table measurements alter column download_jitter_ms drop default;
alter table measurements alter column upload_latency_ms drop default;
alter table measurements alter column upload_jitter_ms drop default;

-- Allow NULL for failed measurements.
alter table measurements alter column download_bps drop not null;
alter table measurements alter column upload_bps drop not null;
alter table measurements alter column latency_ms drop not null;
alter table measurements alter column jitter_ms drop not null;
alter table measurements alter column download_latency_ms drop not null;
alter table measurements alter column download_jitter_ms drop not null;
alter table measurements alter column upload_latency_ms drop not null;
alter table measurements alter column upload_jitter_ms drop not null;

-- Make session_id NOT NULL (app always provides it).
alter table measurements alter column session_id set not null;

-- Backfill: set metrics to NULL for failed measurements where they were 0.
update measurements
set download_bps = null,
    upload_bps = null,
    latency_ms = null,
    jitter_ms = null,
    download_latency_ms = null,
    download_jitter_ms = null,
    upload_latency_ms = null,
    upload_jitter_ms = null
where success = false
  and download_bps = 0
  and upload_bps = 0;

-- Drop the unused index (already done in production, idempotent).
drop index if exists measurements_success_measured_at_idx;
