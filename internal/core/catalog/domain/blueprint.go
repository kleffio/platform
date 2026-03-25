package domain

// ExtensionType is the kind of extensibility slot (mods, plugins, maps, worlds).
type ExtensionType string

const (
	ExtensionTypeMod    ExtensionType = "mod"
	ExtensionTypePlugin ExtensionType = "plugin"
	ExtensionTypeMap    ExtensionType = "map"
	ExtensionTypeWorld  ExtensionType = "world"
)

// SourceEntry is a marketplace or upload source for an extension slot.
type SourceEntry string

// Port is a single port a game server container exposes.
type Port struct {
	ContainerPort int
	Protocol      string // "TCP" or "UDP"
}

// ExtensibilitySlot describes one tab of game-specific content the frontend
// renders. The frontend iterates this array — there is no game-type switch.
type ExtensibilitySlot struct {
	Type              ExtensionType
	Label             string
	Sources           []SourceEntry
	GameVersionEnvKey string
	InstallPath       string
	FileExtension     string
}

// Blueprint is an edition or variant of a Crate (e.g. Minecraft Paper, Bedrock).
// It holds everything needed to provision a game server: image, env defaults,
// ports, and extensibility slots.
type Blueprint struct {
	ID          string
	CrateID     string
	Name        string
	Image       string
	EnvDefaults map[string]string
	Ports       []Port
	Extensibility []ExtensibilitySlot
}
