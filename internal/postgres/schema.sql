create table if not exists measurements (
    id bigserial primary key,
    measured_at timestamptz not null,
    session_id text not null,
    endpoint text not null,
    success boolean not null,
    download_bps double precision,
    upload_bps double precision,
    latency_ms double precision,
    jitter_ms double precision,
    download_latency_ms double precision,
    download_jitter_ms double precision,
    upload_latency_ms double precision,
    upload_jitter_ms double precision,
    raw jsonb not null,
    created_at timestamptz not null default now()
);

create unique index if not exists measurements_session_endpoint_measured_at_idx
    on measurements (session_id, endpoint, measured_at);

create index if not exists measurements_measured_at_idx
    on measurements (measured_at desc);
