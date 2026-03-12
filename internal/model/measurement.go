package model

import "time"

type Measurement struct {
	ID                int64     `json:"id"`
	MeasuredAt        time.Time `json:"measuredAt"`
	SessionID         string    `json:"sessionId"`
	Endpoint          string    `json:"endpoint"`
	Success           bool      `json:"success"`
	DownloadBPS       float64   `json:"downloadBPS"`
	UploadBPS         float64   `json:"uploadBPS"`
	LatencyMS         float64   `json:"latencyMS"`
	JitterMS          float64   `json:"jitterMS"`
	DownloadLatencyMS float64   `json:"downloadLatencyMS"`
	DownloadJitterMS  float64   `json:"downloadJitterMS"`
	UploadLatencyMS   float64   `json:"uploadLatencyMS"`
	UploadJitterMS    float64   `json:"uploadJitterMS"`
	RawJSON           []byte    `json:"raw"`
}
