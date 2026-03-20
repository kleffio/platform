package domain

import "time"

// NodeStatus reflects the operational state of a physical or virtual node.
type NodeStatus string

const (
	NodeStatusOnline      NodeStatus = "online"
	NodeStatusOffline     NodeStatus = "offline"
	NodeStatusDraining    NodeStatus = "draining"
	NodeStatusMaintenance NodeStatus = "maintenance"
)

// Node represents a compute node registered with the platform.
// Nodes are managed by the daemon and reported to the control plane.
type Node struct {
	ID        string
	Hostname  string
	Region    string
	IPAddress string
	Status    NodeStatus

	// Capacity
	TotalVCPU   int
	TotalMemGB  int
	TotalDiskGB int

	// Reported by daemon heartbeat
	UsedVCPU   int
	UsedMemGB  int
	UsedDiskGB int

	LastHeartbeatAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AvailableVCPU returns unallocated CPU.
func (n *Node) AvailableVCPU() int { return n.TotalVCPU - n.UsedVCPU }

// AvailableMemGB returns unallocated memory in GB.
func (n *Node) AvailableMemGB() int { return n.TotalMemGB - n.UsedMemGB }
