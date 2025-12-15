// Package tracker provides beads state tracking and change detection.
package tracker

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Bead represents a single bead snapshot.
type Bead struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Description  string `json:"description"`
	CommentCount int    `json:"comment_count"`
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
	return &Tracker{}
}

// CaptureState parses JSON from bd list --json and returns a State snapshot.
func (t *Tracker) CaptureState(jsonData string) (State, error) {
	var beads []Bead
	if err := json.Unmarshal([]byte(jsonData), &beads); err != nil {
		return State{}, fmt.Errorf("failed to parse beads JSON: %w", err)
	}

	state := State{
		Beads: make(map[string]Bead, len(beads)),
	}
	for _, b := range beads {
		state.Beads[b.ID] = b
	}
	return state, nil
}

// Diff compares two states and returns the changes.
func (t *Tracker) Diff(before, after State) Changes {
	changes := Changes{
		Created:   []Bead{},
		Modified:  []BeadChange{},
		Commented: []BeadChange{},
	}

	// Find created and modified beads
	for id, afterBead := range after.Beads {
		beforeBead, existed := before.Beads[id]
		if !existed {
			// New bead created
			changes.Created = append(changes.Created, afterBead)
			continue
		}

		// Check for modifications
		statusChanged := beforeBead.Status != afterBead.Status
		descChanged := beforeBead.Description != afterBead.Description
		newComments := afterBead.CommentCount - beforeBead.CommentCount

		if statusChanged || descChanged {
			changes.Modified = append(changes.Modified, BeadChange{
				ID:                 id,
				Title:              afterBead.Title,
				OldStatus:          beforeBead.Status,
				NewStatus:          afterBead.Status,
				DescriptionChanged: descChanged,
			})
		}

		if newComments > 0 {
			changes.Commented = append(changes.Commented, BeadChange{
				ID:          id,
				Title:       afterBead.Title,
				NewComments: newComments,
			})
		}
	}

	return changes
}

// FormatSummary returns a human-readable summary of the changes.
func (c Changes) FormatSummary() string {
	if len(c.Created) == 0 && len(c.Modified) == 0 && len(c.Commented) == 0 {
		return "No changes"
	}

	var parts []string

	if len(c.Created) > 0 {
		for _, b := range c.Created {
			parts = append(parts, fmt.Sprintf("Created: %s - %s", b.ID, b.Title))
		}
	}

	if len(c.Modified) > 0 {
		for _, m := range c.Modified {
			if m.OldStatus != m.NewStatus && m.NewStatus != "" {
				parts = append(parts, fmt.Sprintf("Status: %s → %s (%s)", m.OldStatus, m.NewStatus, m.ID))
			}
			if m.DescriptionChanged {
				parts = append(parts, fmt.Sprintf("Updated: %s - %s", m.ID, m.Title))
			}
		}
	}

	if len(c.Commented) > 0 {
		for _, m := range c.Commented {
			parts = append(parts, fmt.Sprintf("Comments: +%d on %s", m.NewComments, m.ID))
		}
	}

	return strings.Join(parts, "\n")
}
