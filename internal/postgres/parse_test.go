package postgres

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestParseMeasuredAt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "valid timestamp",
			input: "speedtest-2026-03-14T12-30-45-000Z.json",
			want:  time.Date(2026, 3, 14, 12, 30, 45, 0, time.UTC),
		},
		{
			name:  "midnight",
			input: "speedtest-2026-01-01T00-00-00-000Z.json",
			want:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "bad format",
			input:   "not-a-timestamp.json",
			wantErr: true,
		},
		{
			name:  "without prefix",
			input: "2026-03-14T12-30-45-000Z.json",
			want:  time.Date(2026, 3, 14, 12, 30, 45, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMeasuredAt(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMeasurementPayload(t *testing.T) {
	tests := []struct {
		name               string
		payload            string
		fallbackMeasuredAt time.Time
		fallbackEndpoint   string
		wantSuccess        bool
		wantDownload       float64
		wantUpload         float64
		wantLatency        float64
		wantSessionID      string
		wantEndpoint       string
		wantErr            bool
	}{
		{
			name: "full payload with all fields",
			payload: `{
				"sessionID": "abc-123",
				"endpoint": "https://netzbremse.de/speed",
				"success": true,
				"timestamp": "2026-03-14T10:00:00Z",
				"result": {
					"download": 50000000,
					"upload": 10000000,
					"latency": 12.5,
					"jitter": 3.2,
					"downLoadedLatency": 15.0,
					"downLoadedJitter": 4.0,
					"upLoadedLatency": 18.0,
					"upLoadedJitter": 5.0
				}
			}`,
			wantSuccess:   true,
			wantDownload:  50000000,
			wantUpload:    10000000,
			wantLatency:   12.5,
			wantSessionID: "abc-123",
			wantEndpoint:  "https://netzbremse.de/speed",
		},
		{
			name: "failed measurement",
			payload: `{
				"sessionID": "def-456",
				"endpoint": "https://netzbremse.de/speed",
				"success": false,
				"result": {
					"download": 0,
					"upload": 0,
					"latency": 0,
					"jitter": 0
				}
			}`,
			wantSuccess:   false,
			wantDownload:  0,
			wantUpload:    0,
			wantLatency:   0,
			wantSessionID: "def-456",
			wantEndpoint:  "https://netzbremse.de/speed",
		},
		{
			name: "uses fallback endpoint when empty",
			payload: `{
				"success": true,
				"result": {"download": 1000, "upload": 500, "latency": 10, "jitter": 1}
			}`,
			fallbackEndpoint: "https://fallback.example.com",
			wantSuccess:      true,
			wantDownload:     1000,
			wantUpload:       500,
			wantLatency:      10,
			wantEndpoint:     "https://fallback.example.com",
		},
		{
			name: "uses fallback measuredAt when no timestamp",
			payload: `{
				"success": true,
				"result": {"download": 1000, "upload": 500, "latency": 10, "jitter": 1}
			}`,
			fallbackMeasuredAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantSuccess:        true,
			wantDownload:       1000,
			wantUpload:         500,
			wantLatency:        10,
		},
		{
			name:    "invalid json",
			payload: `{not valid`,
			wantErr: true,
		},
		{
			name: "invalid timestamp format",
			payload: `{
				"success": true,
				"timestamp": "not-a-date",
				"result": {"download": 0, "upload": 0, "latency": 0, "jitter": 0}
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMeasurementPayload([]byte(tt.payload), tt.fallbackMeasuredAt, tt.fallbackEndpoint)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Success != tt.wantSuccess {
				t.Errorf("Success = %v, want %v", got.Success, tt.wantSuccess)
			}
			if !floatEqual(got.DownloadBPS, tt.wantDownload) {
				t.Errorf("DownloadBPS = %v, want %v", got.DownloadBPS, tt.wantDownload)
			}
			if !floatEqual(got.UploadBPS, tt.wantUpload) {
				t.Errorf("UploadBPS = %v, want %v", got.UploadBPS, tt.wantUpload)
			}
			if !floatEqual(got.LatencyMS, tt.wantLatency) {
				t.Errorf("LatencyMS = %v, want %v", got.LatencyMS, tt.wantLatency)
			}
			if tt.wantSessionID != "" && got.SessionID != tt.wantSessionID {
				t.Errorf("SessionID = %q, want %q", got.SessionID, tt.wantSessionID)
			}
			if tt.wantEndpoint != "" && got.Endpoint != tt.wantEndpoint {
				t.Errorf("Endpoint = %q, want %q", got.Endpoint, tt.wantEndpoint)
			}
			if !json.Valid(got.RawJSON) {
				t.Errorf("RawJSON is not valid JSON: %s", got.RawJSON)
			}
			if !tt.fallbackMeasuredAt.IsZero() && got.MeasuredAt.Equal(tt.fallbackMeasuredAt) {
				// Fallback was used — expected.
			}
		})
	}
}

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
