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

type speedtestDocument struct {
	SessionID string `json:"sessionID"`
	Endpoint  string `json:"endpoint"`
	Success   bool   `json:"success"`
	Result    struct {
		Download struct {
			Bandwidth float64 `json:"bandwidth"`
			Latency   struct {
				Low float64 `json:"low"`
				IQM float64 `json:"iqm"`
			} `json:"latency"`
		} `json:"download"`
		Upload struct {
			Bandwidth float64 `json:"bandwidth"`
			Latency   struct {
				Low float64 `json:"low"`
				IQM float64 `json:"iqm"`
			} `json:"latency"`
		} `json:"upload"`
		Ping struct {
			Latency float64 `json:"latency"`
			Jitter  float64 `json:"jitter"`
		} `json:"ping"`
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
			coalesce(session_id::text, ''),
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
		item.RawJSON = []byte(raw)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate measurements: %w", err)
	}
	return result, nil
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
		item, err := parseMeasurement(path, payload)
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

func parseMeasurement(path string, payload []byte) (model.Measurement, error) {
	var doc speedtestDocument
	if err := json.Unmarshal(payload, &doc); err != nil {
		return model.Measurement{}, fmt.Errorf("decode json: %w", err)
	}

	measuredAt, err := parseMeasuredAt(filepath.Base(path))
	if err != nil {
		return model.Measurement{}, err
	}

	return model.Measurement{
		MeasuredAt:        measuredAt,
		SessionID:         doc.SessionID,
		Endpoint:          doc.Endpoint,
		Success:           doc.Success,
		DownloadBPS:       doc.Result.Download.Bandwidth,
		UploadBPS:         doc.Result.Upload.Bandwidth,
		LatencyMS:         doc.Result.Ping.Latency,
		JitterMS:          doc.Result.Ping.Jitter,
		DownloadLatencyMS: doc.Result.Download.Latency.Low,
		DownloadJitterMS:  doc.Result.Download.Latency.IQM,
		UploadLatencyMS:   doc.Result.Upload.Latency.Low,
		UploadJitterMS:    doc.Result.Upload.Latency.IQM,
		RawJSON:           payload,
	}, nil
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
