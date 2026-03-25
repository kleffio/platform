package domain

// Crate represents a game (e.g. Minecraft, Rust).
// A crate groups one or more Blueprints (editions/variants).
type Crate struct {
	ID          string
	Name        string
	Description string
	ImageURL    string
}
