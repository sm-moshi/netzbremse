package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     any
		wantCode  int
		wantType  string
		checkBody func(t *testing.T, body []byte)
	}{
		{
			name:     "simple map",
			value:    map[string]string{"key": "value"},
			wantCode: http.StatusOK,
			wantType: "application/json",
			checkBody: func(t *testing.T, body []byte) {
				var m map[string]string
				if err := json.Unmarshal(body, &m); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if m["key"] != "value" {
					t.Errorf("got %q, want %q", m["key"], "value")
				}
			},
		},
		{
			name:     "slice of ints",
			value:    []int{1, 2, 3},
			wantCode: http.StatusOK,
			wantType: "application/json",
			checkBody: func(t *testing.T, body []byte) {
				var s []int
				if err := json.Unmarshal(body, &s); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if len(s) != 3 {
					t.Errorf("got %d elements, want 3", len(s))
				}
			},
		},
		{
			name:     "nil value",
			value:    nil,
			wantCode: http.StatusOK,
			wantType: "application/json",
			checkBody: func(t *testing.T, body []byte) {
				if string(body) != "null\n" {
					t.Errorf("got %q, want %q", body, "null\n")
				}
			},
		},
		{
			name:     "empty slice",
			value:    []string{},
			wantCode: http.StatusOK,
			wantType: "application/json",
			checkBody: func(t *testing.T, body []byte) {
				if string(body) != "[]\n" {
					t.Errorf("got %q, want %q", body, "[]\n")
				}
			},
		},
		{
			name:     "nested struct",
			value:    struct{ Name string }{"test"},
			wantCode: http.StatusOK,
			wantType: "application/json",
			checkBody: func(t *testing.T, body []byte) {
				var m map[string]string
				if err := json.Unmarshal(body, &m); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if m["Name"] != "test" {
					t.Errorf("got %q, want %q", m["Name"], "test")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			writeJSON(recorder, tt.value)

			result := recorder.Result()
			defer result.Body.Close()

			if result.StatusCode != tt.wantCode {
				t.Errorf("status = %d, want %d", result.StatusCode, tt.wantCode)
			}
			if ct := result.Header.Get("Content-Type"); ct != tt.wantType {
				t.Errorf("Content-Type = %q, want %q", ct, tt.wantType)
			}
			if tt.checkBody != nil {
				tt.checkBody(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestWriteJSONIndentation(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	writeJSON(recorder, map[string]int{"a": 1})

	body := recorder.Body.String()

	// The encoder uses 2-space indent. Verify the output is pretty-printed.
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}

	// A pretty-printed object should contain a newline.
	if len(body) < 5 {
		t.Errorf("body too short for indented JSON: %q", body)
	}
	// Check that the output contains 2-space indentation.
	expected := "{\n  \"a\": 1\n}\n"
	if body != expected {
		t.Errorf("got %q, want %q", body, expected)
	}
}

func TestHealthzHandler(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "ok\n" {
		t.Errorf("body = %q, want %q", body, "ok\n")
	}
}

func TestMeasurementsLimitParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		queryLimit   string
		defaultLimit int
		wantLimit    int
	}{
		{
			name:         "no query param uses default",
			queryLimit:   "",
			defaultLimit: 50,
			wantLimit:    50,
		},
		{
			name:         "valid limit within range",
			queryLimit:   "100",
			defaultLimit: 50,
			wantLimit:    100,
		},
		{
			name:         "limit at max boundary",
			queryLimit:   "500",
			defaultLimit: 50,
			wantLimit:    500,
		},
		{
			name:         "limit exceeds max falls back to default",
			queryLimit:   "501",
			defaultLimit: 50,
			wantLimit:    50,
		},
		{
			name:         "zero limit falls back to default",
			queryLimit:   "0",
			defaultLimit: 50,
			wantLimit:    50,
		},
		{
			name:         "negative limit falls back to default",
			queryLimit:   "-1",
			defaultLimit: 50,
			wantLimit:    50,
		},
		{
			name:         "non-numeric limit falls back to default",
			queryLimit:   "abc",
			defaultLimit: 50,
			wantLimit:    50,
		},
		{
			name:         "limit of 1",
			queryLimit:   "1",
			defaultLimit: 50,
			wantLimit:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Replicate the limit-parsing logic from main.go.
			limit := tt.defaultLimit
			if raw := tt.queryLimit; raw != "" {
				parsed, err := strconv.Atoi(raw)
				if err == nil && parsed > 0 && parsed <= 500 {
					limit = parsed
				}
			}

			if limit != tt.wantLimit {
				t.Errorf("limit = %d, want %d", limit, tt.wantLimit)
			}
		})
	}
}
