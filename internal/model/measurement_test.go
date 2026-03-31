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
		DownloadBPS:       Float64Ptr(50_000_000),
		UploadBPS:         Float64Ptr(10_000_000),
		LatencyMS:         Float64Ptr(12.5),
		JitterMS:          Float64Ptr(3.2),
		DownloadLatencyMS: Float64Ptr(15.0),
		DownloadJitterMS:  Float64Ptr(4.0),
		UploadLatencyMS:   Float64Ptr(18.0),
		UploadJitterMS:    Float64Ptr(5.0),
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
	assertFloat64PtrEqual(t, "DownloadBPS", decoded.DownloadBPS, original.DownloadBPS)
	assertFloat64PtrEqual(t, "UploadBPS", decoded.UploadBPS, original.UploadBPS)
	assertFloat64PtrEqual(t, "LatencyMS", decoded.LatencyMS, original.LatencyMS)
	assertFloat64PtrEqual(t, "JitterMS", decoded.JitterMS, original.JitterMS)
	assertFloat64PtrEqual(t, "DownloadLatencyMS", decoded.DownloadLatencyMS, original.DownloadLatencyMS)
	assertFloat64PtrEqual(t, "DownloadJitterMS", decoded.DownloadJitterMS, original.DownloadJitterMS)
	assertFloat64PtrEqual(t, "UploadLatencyMS", decoded.UploadLatencyMS, original.UploadLatencyMS)
	assertFloat64PtrEqual(t, "UploadJitterMS", decoded.UploadJitterMS, original.UploadJitterMS)
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
		DownloadBPS:       Float64Ptr(100),
		UploadBPS:         Float64Ptr(50),
		LatencyMS:         Float64Ptr(10),
		JitterMS:          Float64Ptr(2),
		DownloadLatencyMS: Float64Ptr(11),
		DownloadJitterMS:  Float64Ptr(3),
		UploadLatencyMS:   Float64Ptr(12),
		UploadJitterMS:    Float64Ptr(4),
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
	if decoded.DownloadBPS != nil {
		t.Errorf("DownloadBPS: got %v, want nil", decoded.DownloadBPS)
	}
	if decoded.SessionID != "" {
		t.Errorf("SessionID: got %q, want empty", decoded.SessionID)
	}
}

func TestMeasurementNullMetrics(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": 1,
		"measuredAt": "2026-01-01T00:00:00Z",
		"sessionId": "s1",
		"endpoint": "ep",
		"success": false,
		"downloadBPS": null,
		"uploadBPS": null,
		"latencyMS": null,
		"jitterMS": null,
		"downloadLatencyMS": null,
		"downloadJitterMS": null,
		"uploadLatencyMS": null,
		"uploadJitterMS": null,
		"raw": {}
	}`

	var m Measurement
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m.DownloadBPS != nil {
		t.Errorf("DownloadBPS: got %v, want nil", m.DownloadBPS)
	}
	if m.UploadBPS != nil {
		t.Errorf("UploadBPS: got %v, want nil", m.UploadBPS)
	}
	if m.Success {
		t.Error("Success: got true, want false")
	}
}

func TestMeasurementUnmarshalFromExternalJSON(t *testing.T) {
	t.Parallel()

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
	if m.DownloadBPS == nil || *m.DownloadBPS != 75000000.5 {
		t.Errorf("DownloadBPS: got %v, want 75000000.5", m.DownloadBPS)
	}
	if !m.Success {
		t.Error("Success: got false, want true")
	}
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
		"downloadBPS": null,
		"uploadBPS": null,
		"latencyMS": null,
		"jitterMS": null,
		"downloadLatencyMS": null,
		"downloadJitterMS": null,
		"uploadLatencyMS": null,
		"uploadJitterMS": null,
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

func assertFloat64PtrEqual(t *testing.T, name string, got, want *float64) {
	t.Helper()
	if got == nil && want == nil {
		return
	}
	if got == nil || want == nil {
		t.Errorf("%s: got %v, want %v", name, got, want)
		return
	}
	if *got != *want {
		t.Errorf("%s: got %f, want %f", name, *got, *want)
	}
}
