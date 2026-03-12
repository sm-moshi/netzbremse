package model

import "time"

type Measurement struct {
	ID                int64
	MeasuredAt        time.Time
	SessionID         string
	Endpoint          string
	Success           bool
	DownloadBPS       float64
	UploadBPS         float64
	LatencyMS         float64
	JitterMS          float64
	DownloadLatencyMS float64
	DownloadJitterMS  float64
	UploadLatencyMS   float64
	UploadJitterMS    float64
	RawJSON           []byte
}
