package domain

import "time"

// LogLine is one line of output from a running workload container.
type LogLine struct {
	ID         int64     `json:"id"`
	WorkloadID string    `json:"workload_id"`
	ProjectID  string    `json:"project_id"`
	Ts         time.Time `json:"ts"`
	Stream     string    `json:"stream"`
	Line       string    `json:"line"`
}
