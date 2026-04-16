package ports

// WorkloadSpec is the daemon queue payload for workload operations.
type WorkloadSpec struct {
	OwnerID          string            `json:"owner_id"`
	ServerID         string            `json:"server_id"`
	BlueprintID      string            `json:"blueprint_id"`
	ProjectID        string            `json:"project_id"`
	ProjectSlug      string            `json:"project_slug"`
	Image            string            `json:"image"`
	BlueprintVersion string            `json:"blueprint_version,omitempty"`
	EnvOverrides     map[string]string `json:"env_overrides,omitempty"`
	MemoryBytes      int64             `json:"memory_bytes,omitempty"`
	CPUMillicores    int64             `json:"cpu_millicores,omitempty"`
	PortRequirements []PortRequirement `json:"port_requirements,omitempty"`
	RuntimeHints     RuntimeHints      `json:"runtime_hints,omitempty"`
}

type PortRequirement struct {
	TargetPort int    `json:"target_port"`
	Protocol   string `json:"protocol"`
}

type RuntimeHints struct {
	KubernetesStrategy string `json:"kubernetes_strategy,omitempty"`
	ExposeUDP          bool   `json:"expose_udp,omitempty"`
	PersistentStorage  bool   `json:"persistent_storage,omitempty"`
	StoragePath        string `json:"storage_path,omitempty"`
	StorageGB          int    `json:"storage_gb,omitempty"`
}
