package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMeasurementJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := Measurement{
		ID:                42,
		MeasuredAt:        time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC),
		SessionID:         "sess-abc-123",
		Endpoint:          "https://netzbremse.de/speed",
		Success:           true,
		DownloadBPS:       50_000_000,
		UploadBPS:         10_000_000,
		LatencyMS:         12.5,
		JitterMS:          3.2,
		DownloadLatencyMS: 15.0,
		DownloadJitterMS:  4.0,
		UploadLatencyMS:   18.0,
		UploadJitterMS:    5.0,
		RawJSON:           json.RawMessage(`{"extra":"data"}`),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Measurement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, original.ID)
	}
	if !decoded.MeasuredAt.Equal(original.MeasuredAt) {
		t.Errorf("MeasuredAt: got %v, want %v", decoded.MeasuredAt, original.MeasuredAt)
	}
	if decoded.SessionID != original.SessionID {
		t.Errorf("SessionID: got %q, want %q", decoded.SessionID, original.SessionID)
	}
	if decoded.Endpoint != original.Endpoint {
		t.Errorf("Endpoint: got %q, want %q", decoded.Endpoint, original.Endpoint)
	}
	if decoded.Success != original.Success {
		t.Errorf("Success: got %v, want %v", decoded.Success, original.Success)
	}
	if decoded.DownloadBPS != original.DownloadBPS {
		t.Errorf("DownloadBPS: got %f, want %f", decoded.DownloadBPS, original.DownloadBPS)
	}
	if decoded.UploadBPS != original.UploadBPS {
		t.Errorf("UploadBPS: got %f, want %f", decoded.UploadBPS, original.UploadBPS)
	}
	if decoded.LatencyMS != original.LatencyMS {
		t.Errorf("LatencyMS: got %f, want %f", decoded.LatencyMS, original.LatencyMS)
	}
	if decoded.JitterMS != original.JitterMS {
		t.Errorf("JitterMS: got %f, want %f", decoded.JitterMS, original.JitterMS)
	}
	if decoded.DownloadLatencyMS != original.DownloadLatencyMS {
		t.Errorf("DownloadLatencyMS: got %f, want %f", decoded.DownloadLatencyMS, original.DownloadLatencyMS)
	}
	if decoded.DownloadJitterMS != original.DownloadJitterMS {
		t.Errorf("DownloadJitterMS: got %f, want %f", decoded.DownloadJitterMS, original.DownloadJitterMS)
	}
	if decoded.UploadLatencyMS != original.UploadLatencyMS {
		t.Errorf("UploadLatencyMS: got %f, want %f", decoded.UploadLatencyMS, original.UploadLatencyMS)
	}
	if decoded.UploadJitterMS != original.UploadJitterMS {
		t.Errorf("UploadJitterMS: got %f, want %f", decoded.UploadJitterMS, original.UploadJitterMS)
	}
	if string(decoded.RawJSON) != string(original.RawJSON) {
		t.Errorf("RawJSON: got %s, want %s", decoded.RawJSON, original.RawJSON)
	}
}

func TestMeasurementJSONFieldNames(t *testing.T) {
	t.Parallel()

	m := Measurement{
		ID:                1,
		MeasuredAt:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		SessionID:         "s1",
		Endpoint:          "ep",
		Success:           true,
		DownloadBPS:       100,
		UploadBPS:         50,
		LatencyMS:         10,
		JitterMS:          2,
		DownloadLatencyMS: 11,
		DownloadJitterMS:  3,
		UploadLatencyMS:   12,
		UploadJitterMS:    4,
		RawJSON:           json.RawMessage(`{}`),
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	expectedKeys := []string{
		"id", "measuredAt", "sessionId", "endpoint", "success",
		"downloadBPS", "uploadBPS", "latencyMS", "jitterMS",
		"downloadLatencyMS", "downloadJitterMS",
		"uploadLatencyMS", "uploadJitterMS", "raw",
	}

	for _, key := range expectedKeys {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing JSON key %q in marshalled output", key)
		}
	}

	if len(fields) != len(expectedKeys) {
		t.Errorf("expected %d JSON keys, got %d", len(expectedKeys), len(fields))
	}
}

func TestMeasurementZeroValue(t *testing.T) {
	t.Parallel()

	var m Measurement

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal zero value: %v", err)
	}

	var decoded Measurement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal zero value: %v", err)
	}

	if decoded.ID != 0 {
		t.Errorf("ID: got %d, want 0", decoded.ID)
	}
	if decoded.Success {
		t.Error("Success: got true, want false")
	}
	if decoded.DownloadBPS != 0 {
		t.Errorf("DownloadBPS: got %f, want 0", decoded.DownloadBPS)
	}
	if decoded.SessionID != "" {
		t.Errorf("SessionID: got %q, want empty", decoded.SessionID)
	}
}

func TestMeasurementUnmarshalFromExternalJSON(t *testing.T) {
	t.Parallel()

	// Simulate the kind of JSON the dashboard API might return.
	raw := `{
		"id": 99,
		"measuredAt": "2026-03-15T08:00:00Z",
		"sessionId": "ext-session",
		"endpoint": "https://example.com",
		"success": true,
		"downloadBPS": 75000000.5,
		"uploadBPS": 25000000.25,
		"latencyMS": 8.75,
		"jitterMS": 1.5,
		"downloadLatencyMS": 9.0,
		"downloadJitterMS": 2.0,
		"uploadLatencyMS": 10.0,
		"uploadJitterMS": 3.0,
		"raw": {"nested": "value"}
	}`

	var m Measurement
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m.ID != 99 {
		t.Errorf("ID: got %d, want 99", m.ID)
	}
	if m.SessionID != "ext-session" {
		t.Errorf("SessionID: got %q, want %q", m.SessionID, "ext-session")
	}
	if m.DownloadBPS != 75000000.5 {
		t.Errorf("DownloadBPS: got %f, want 75000000.5", m.DownloadBPS)
	}
	if !m.Success {
		t.Error("Success: got false, want true")
	}

	// RawJSON should preserve the nested object.
	if !json.Valid(m.RawJSON) {
		t.Errorf("RawJSON is not valid JSON: %s", m.RawJSON)
	}
}

func TestMeasurementNullRawJSON(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": 1,
		"measuredAt": "2026-01-01T00:00:00Z",
		"sessionId": "",
		"endpoint": "",
		"success": false,
		"downloadBPS": 0,
		"uploadBPS": 0,
		"latencyMS": 0,
		"jitterMS": 0,
		"downloadLatencyMS": 0,
		"downloadJitterMS": 0,
		"uploadLatencyMS": 0,
		"uploadJitterMS": 0,
		"raw": null
	}`

	var m Measurement
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal with null raw: %v", err)
	}

	if m.RawJSON != nil && string(m.RawJSON) != "null" {
		t.Errorf("RawJSON: got %s, want nil or null", m.RawJSON)
	}
}
