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
	GameServerID   string
	NodeID         string
	RecordedAt     time.Time

	CPUSeconds     float64
	MemoryGBHours  float64
	NetworkInMB    float64
	NetworkOutMB   float64
	DiskReadMB     float64
	DiskWriteMB    float64
}
