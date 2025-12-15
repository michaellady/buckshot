// Package tracker provides beads state tracking and change detection.
package tracker

// Bead represents a single bead snapshot.
type Bead struct {
	ID           string
	Title        string
	Status       string
	Description  string
	CommentCount int
}

// State represents a snapshot of beads at a point in time.
type State struct {
	Beads map[string]Bead
}

// BeadChange represents a change to a single bead.
type BeadChange struct {
	ID                 string
	Title              string
	OldStatus          string
	NewStatus          string
	DescriptionChanged bool
	NewComments        int
}

// Changes represents the diff between two states.
type Changes struct {
	Created   []Bead
	Modified  []BeadChange
	Commented []BeadChange
}

// Tracker tracks beads state changes.
type Tracker struct{}

// NewTracker creates a new Tracker.
func NewTracker() *Tracker {
	// TODO: Implement in GREEN phase
	return nil
}

// CaptureState parses JSON from bd list --json and returns a State snapshot.
func (t *Tracker) CaptureState(jsonData string) (State, error) {
	// TODO: Implement in GREEN phase
	return State{}, nil
}

// Diff compares two states and returns the changes.
func (t *Tracker) Diff(before, after State) Changes {
	// TODO: Implement in GREEN phase
	return Changes{}
}

// FormatSummary returns a human-readable summary of the changes.
func (c Changes) FormatSummary() string {
	// TODO: Implement in GREEN phase
	return ""
}
