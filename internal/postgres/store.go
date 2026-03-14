package postgres

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sm-moshi/netzbremse/internal/model"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	pool *pgxpool.Pool
}

type Overview struct {
	TotalCount      int64     `json:"totalCount"`
	RecentCount     int64     `json:"recentCount"`
	SuccessRate     float64   `json:"successRate"`
	AverageDownload float64   `json:"averageDownload"`
	AverageUpload   float64   `json:"averageUpload"`
	AverageLatency  float64   `json:"averageLatency"`
	LastMeasuredAt  time.Time `json:"lastMeasuredAt"`
	LastEndpoint    string    `json:"lastEndpoint"`
}

type speedtestDocument struct {
	SessionID string `json:"sessionID"`
	Endpoint  string `json:"endpoint"`
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
	Result    struct {
		Download        float64 `json:"download"`
		Upload          float64 `json:"upload"`
		Latency         float64 `json:"latency"`
		Jitter          float64 `json:"jitter"`
		DownloadLatency float64 `json:"downLoadedLatency"`
		DownloadJitter  float64 `json:"downLoadedJitter"`
		UploadLatency   float64 `json:"upLoadedLatency"`
		UploadJitter    float64 `json:"upLoadedJitter"`
	} `json:"result"`
}

func New(ctx context.Context, uri string) (*Store, error) {
	pool, err := pgxpool.New(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if _, err := s.pool.Exec(ctx, `
		do $$
		begin
			if exists (
				select 1
				from information_schema.columns
				where table_schema = 'public'
					and table_name = 'measurements'
					and column_name = 'session_id'
					and udt_name = 'uuid'
			) then
				alter table measurements
					alter column session_id type text
					using session_id::text;
			end if;
		end
		$$;
	`); err != nil {
		return fmt.Errorf("migrate session_id column: %w", err)
	}
	return nil
}

func (s *Store) Insert(ctx context.Context, measurement model.Measurement) error {
	_, err := s.pool.Exec(ctx, `
		insert into measurements (
			measured_at,
			session_id,
			endpoint,
			success,
			download_bps,
			upload_bps,
			latency_ms,
			jitter_ms,
			download_latency_ms,
			download_jitter_ms,
			upload_latency_ms,
			upload_jitter_ms,
			raw
		) values (
			$1, nullif($2, ''), $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb
		)
		on conflict do nothing
	`,
		measurement.MeasuredAt,
		measurement.SessionID,
		measurement.Endpoint,
		measurement.Success,
		measurement.DownloadBPS,
		measurement.UploadBPS,
		measurement.LatencyMS,
		measurement.JitterMS,
		measurement.DownloadLatencyMS,
		measurement.DownloadJitterMS,
		measurement.UploadLatencyMS,
		measurement.UploadJitterMS,
		string(measurement.RawJSON),
	)
	if err != nil {
		return fmt.Errorf("insert measurement: %w", err)
	}
	return nil
}

func (s *Store) ListLatest(ctx context.Context, limit int) ([]model.Measurement, error) {
	rows, err := s.pool.Query(ctx, `
		select
			id,
			measured_at,
			coalesce(session_id, ''),
			endpoint,
			success,
			download_bps,
			upload_bps,
			latency_ms,
			jitter_ms,
			download_latency_ms,
			download_jitter_ms,
			upload_latency_ms,
			upload_jitter_ms,
			raw::text
		from measurements
		order by measured_at desc
		limit $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query latest measurements: %w", err)
	}
	defer rows.Close()

	result := make([]model.Measurement, 0, limit)
	for rows.Next() {
		var item model.Measurement
		var raw string
		if err := rows.Scan(
			&item.ID,
			&item.MeasuredAt,
			&item.SessionID,
			&item.Endpoint,
			&item.Success,
			&item.DownloadBPS,
			&item.UploadBPS,
			&item.LatencyMS,
			&item.JitterMS,
			&item.DownloadLatencyMS,
			&item.DownloadJitterMS,
			&item.UploadLatencyMS,
			&item.UploadJitterMS,
			&raw,
		); err != nil {
			return nil, fmt.Errorf("scan measurement: %w", err)
		}
		item.RawJSON = json.RawMessage(raw)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate measurements: %w", err)
	}
	return result, nil
}

func (s *Store) LoadOverview(ctx context.Context) (Overview, error) {
	var overview Overview
	err := s.pool.QueryRow(ctx, `
		with recent as (
			select *
			from measurements
			where measured_at >= now() - interval '24 hours'
		),
		last_row as (
			select measured_at, endpoint
			from measurements
			order by measured_at desc
			limit 1
		)
		select
			coalesce((select count(*) from measurements), 0) as total_count,
			coalesce((select count(*) from recent), 0) as recent_count,
			coalesce((select avg(case when success then 1.0 else 0.0 end) from recent), 0) as success_rate,
			coalesce((select avg(download_bps) from recent where success), 0) as average_download,
			coalesce((select avg(upload_bps) from recent where success), 0) as average_upload,
			coalesce((select avg(latency_ms) from recent where success), 0) as average_latency,
			coalesce((select measured_at from last_row), to_timestamp(0)) as last_measured_at,
			coalesce((select endpoint from last_row), '') as last_endpoint
	`).Scan(
		&overview.TotalCount,
		&overview.RecentCount,
		&overview.SuccessRate,
		&overview.AverageDownload,
		&overview.AverageUpload,
		&overview.AverageLatency,
		&overview.LastMeasuredAt,
		&overview.LastEndpoint,
	)
	if err != nil {
		return Overview{}, fmt.Errorf("load overview: %w", err)
	}
	return overview, nil
}

func (s *Store) ImportDir(ctx context.Context, dir string) (int, error) {
	if dir == "" {
		return 0, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read import dir %q: %w", dir, err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(paths)

	imported := 0
	for _, path := range paths {
		payload, err := os.ReadFile(path)
		if err != nil {
			return imported, fmt.Errorf("read %q: %w", path, err)
		}
		item, err := ParseMeasurementFile(path, payload, "")
		if err != nil {
			return imported, fmt.Errorf("parse %q: %w", path, err)
		}
		if err := s.Insert(ctx, item); err != nil {
			return imported, fmt.Errorf("insert %q: %w", path, err)
		}
		imported++
	}

	return imported, nil
}

func ParseMeasurementPayload(payload []byte, fallbackMeasuredAt time.Time, fallbackEndpoint string) (model.Measurement, error) {
	var doc speedtestDocument
	if err := json.Unmarshal(payload, &doc); err != nil {
		return model.Measurement{}, fmt.Errorf("decode json: %w", err)
	}

	measuredAt := fallbackMeasuredAt
	if doc.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339Nano, doc.Timestamp)
		if err != nil {
			return model.Measurement{}, fmt.Errorf("parse timestamp: %w", err)
		}
		measuredAt = parsed.UTC()
	}
	if measuredAt.IsZero() {
		measuredAt = time.Now().UTC()
	}

	endpoint := doc.Endpoint
	if endpoint == "" {
		endpoint = fallbackEndpoint
	}

	return model.Measurement{
		MeasuredAt:        measuredAt,
		SessionID:         doc.SessionID,
		Endpoint:          endpoint,
		Success:           doc.Success,
		DownloadBPS:       doc.Result.Download,
		UploadBPS:         doc.Result.Upload,
		LatencyMS:         doc.Result.Latency,
		JitterMS:          doc.Result.Jitter,
		DownloadLatencyMS: doc.Result.DownloadLatency,
		DownloadJitterMS:  doc.Result.DownloadJitter,
		UploadLatencyMS:   doc.Result.UploadLatency,
		UploadJitterMS:    doc.Result.UploadJitter,
		RawJSON:           payload,
	}, nil
}

func ParseMeasurementFile(path string, payload []byte, fallbackEndpoint string) (model.Measurement, error) {
	measuredAt, err := parseMeasuredAt(filepath.Base(path))
	if err != nil {
		return model.Measurement{}, err
	}
	return ParseMeasurementPayload(payload, measuredAt, fallbackEndpoint)
}

func parseMeasuredAt(name string) (time.Time, error) {
	stamp := strings.TrimPrefix(name, "speedtest-")
	stamp = strings.TrimSuffix(stamp, ".json")
	parsed, err := time.Parse("2006-01-02T15-04-05-000Z", stamp)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp from %q: %w", name, err)
	}
	return parsed.UTC(), nil
}
