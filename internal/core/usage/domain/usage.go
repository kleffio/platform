package domain

import "time"

// UsageSummary aggregates metered resource consumption for a billing period.
type UsageSummary struct {
	OrganizationID string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	ServerHours    float64
	BandwidthGB    float64
	StorageGB      float64
	EstimatedCost  int // cents
}

// UsageRecord is a raw metering event produced by daemon heartbeats.
type UsageRecord struct {
	ID             string
	OrganizationID string
	ProjectID      string
	GameServerID   string
	NodeID         string
	RecordedAt     time.Time

	// Billing units
	CPUSeconds    float64
	MemoryGBHours float64
	NetworkInMB   float64
	NetworkOutMB  float64
	DiskReadMB    float64
	DiskWriteMB   float64

	// Display units (human-readable monitoring)
	CPUMillicores   int64
	MemoryMB        int64
	NetworkInKbps   float64
	NetworkOutKbps  float64
	DiskReadKbps    float64
	DiskWriteKbps   float64
}

// WorkloadMetrics is the latest snapshot for a single workload, used by the monitoring page.
type WorkloadMetrics struct {
	WorkloadID    string    `json:"workload_id"`
	ProjectID     string    `json:"project_id"`
	CPUMillicores int64     `json:"cpu_millicores"`
	MemoryMB      int64     `json:"memory_mb"`
	NetworkInKbps float64   `json:"network_in_kbps"`
	NetworkOutKbps float64  `json:"network_out_kbps"`
	DiskReadKbps  float64   `json:"disk_read_kbps"`
	DiskWriteKbps float64   `json:"disk_write_kbps"`
	RecordedAt    time.Time `json:"recorded_at"`

	// Allocation limits from the workload provisioning request.
	CPULimitMillicores int64 `json:"cpu_limit_millicores"`
	MemoryLimitBytes   int64 `json:"memory_limit_bytes"`
}
