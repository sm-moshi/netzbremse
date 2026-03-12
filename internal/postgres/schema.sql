create table if not exists measurements (
    id bigserial primary key,
    measured_at timestamptz not null,
    session_id uuid,
    endpoint text not null,
    success boolean not null,
    download_bps double precision not null default 0,
    upload_bps double precision not null default 0,
    latency_ms double precision not null default 0,
    jitter_ms double precision not null default 0,
    download_latency_ms double precision not null default 0,
    download_jitter_ms double precision not null default 0,
    upload_latency_ms double precision not null default 0,
    upload_jitter_ms double precision not null default 0,
    raw jsonb not null,
    created_at timestamptz not null default now()
);

create unique index if not exists measurements_session_endpoint_measured_at_idx
    on measurements (session_id, endpoint, measured_at);

create index if not exists measurements_measured_at_idx
    on measurements (measured_at desc);

create index if not exists measurements_success_measured_at_idx
    on measurements (success, measured_at desc);
