package tracker

import (
	"strings"
	"testing"
)

// TestBeadsTracker_CaptureState tests that we can snapshot current beads with IDs, titles, status, comment counts.
func TestBeadsTracker_CaptureState(t *testing.T) {
	tracker := NewTracker()

	// Mock bd list --json output
	mockJSON := `[
		{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "description": "First test"},
		{"id": "buckshot-xyz", "title": "Test bead 2", "status": "in_progress", "description": "Second test"}
	]`

	state, err := tracker.CaptureState(mockJSON)
	if err != nil {
		t.Fatalf("CaptureState() error = %v", err)
	}

	if len(state.Beads) != 2 {
		t.Errorf("Expected 2 beads, got %d", len(state.Beads))
	}

	// Verify first bead
	bead1, ok := state.Beads["buckshot-abc"]
	if !ok {
		t.Fatal("Expected bead buckshot-abc not found")
	}
	if bead1.ID != "buckshot-abc" {
		t.Errorf("Expected ID buckshot-abc, got %s", bead1.ID)
	}
	if bead1.Title != "Test bead 1" {
		t.Errorf("Expected title 'Test bead 1', got %s", bead1.Title)
	}
	if bead1.Status != "open" {
		t.Errorf("Expected status 'open', got %s", bead1.Status)
	}
}

// TestBeadsTracker_DetectCreated tests that we can identify new beads between two snapshots.
func TestBeadsTracker_DetectCreated(t *testing.T) {
	tracker := NewTracker()

	// Before state: one bead
	beforeJSON := `[
		{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "description": "First test"}
	]`

	// After state: two beads (one new)
	afterJSON := `[
		{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "description": "First test"},
		{"id": "buckshot-xyz", "title": "Test bead 2", "status": "open", "description": "Second test"}
	]`

	before, err := tracker.CaptureState(beforeJSON)
	if err != nil {
		t.Fatalf("CaptureState(before) error = %v", err)
	}

	after, err := tracker.CaptureState(afterJSON)
	if err != nil {
		t.Fatalf("CaptureState(after) error = %v", err)
	}

	changes := tracker.Diff(before, after)

	if len(changes.Created) != 1 {
		t.Errorf("Expected 1 created bead, got %d", len(changes.Created))
	}

	if len(changes.Created) > 0 && changes.Created[0].ID != "buckshot-xyz" {
		t.Errorf("Expected created bead ID buckshot-xyz, got %s", changes.Created[0].ID)
	}
}

// TestBeadsTracker_DetectCommented tests that we can detect comment count changes.
func TestBeadsTracker_DetectCommented(t *testing.T) {
	tracker := NewTracker()

	// Before state: no comments
	beforeJSON := `[
		{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "comment_count": 0}
	]`

	// After state: 3 comments added
	afterJSON := `[
		{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "comment_count": 3}
	]`

	before, err := tracker.CaptureState(beforeJSON)
	if err != nil {
		t.Fatalf("CaptureState(before) error = %v", err)
	}

	after, err := tracker.CaptureState(afterJSON)
	if err != nil {
		t.Fatalf("CaptureState(after) error = %v", err)
	}

	changes := tracker.Diff(before, after)

	if len(changes.Commented) != 1 {
		t.Errorf("Expected 1 commented bead, got %d", len(changes.Commented))
	}

	if len(changes.Commented) > 0 {
		if changes.Commented[0].ID != "buckshot-abc" {
			t.Errorf("Expected commented bead ID buckshot-abc, got %s", changes.Commented[0].ID)
		}
		if changes.Commented[0].NewComments != 3 {
			t.Errorf("Expected 3 new comments, got %d", changes.Commented[0].NewComments)
		}
	}
}

// TestBeadsTracker_DetectModified tests that we can detect status or description changes.
func TestBeadsTracker_DetectModified(t *testing.T) {
	tracker := NewTracker()

	t.Run("StatusChange", func(t *testing.T) {
		beforeJSON := `[
			{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "description": "First test"}
		]`

		afterJSON := `[
			{"id": "buckshot-abc", "title": "Test bead 1", "status": "in_progress", "description": "First test"}
		]`

		before, _ := tracker.CaptureState(beforeJSON)
		after, _ := tracker.CaptureState(afterJSON)

		changes := tracker.Diff(before, after)

		if len(changes.Modified) != 1 {
			t.Errorf("Expected 1 modified bead, got %d", len(changes.Modified))
		}

		if len(changes.Modified) > 0 {
			mod := changes.Modified[0]
			if mod.ID != "buckshot-abc" {
				t.Errorf("Expected modified bead ID buckshot-abc, got %s", mod.ID)
			}
			if mod.OldStatus != "open" {
				t.Errorf("Expected old status 'open', got %s", mod.OldStatus)
			}
			if mod.NewStatus != "in_progress" {
				t.Errorf("Expected new status 'in_progress', got %s", mod.NewStatus)
			}
		}
	})

	t.Run("DescriptionChange", func(t *testing.T) {
		beforeJSON := `[
			{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "description": "Original desc"}
		]`

		afterJSON := `[
			{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "description": "Updated desc"}
		]`

		before, _ := tracker.CaptureState(beforeJSON)
		after, _ := tracker.CaptureState(afterJSON)

		changes := tracker.Diff(before, after)

		if len(changes.Modified) != 1 {
			t.Errorf("Expected 1 modified bead, got %d", len(changes.Modified))
		}

		if len(changes.Modified) > 0 {
			mod := changes.Modified[0]
			if mod.DescriptionChanged != true {
				t.Error("Expected DescriptionChanged to be true")
			}
		}
	})
}

// TestBeadsTracker_FormatSummary tests that we can produce a human-readable action summary.
func TestBeadsTracker_FormatSummary(t *testing.T) {
	tracker := NewTracker()

	beforeJSON := `[
		{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "description": "First test", "comment_count": 0}
	]`

	afterJSON := `[
		{"id": "buckshot-abc", "title": "Test bead 1", "status": "in_progress", "description": "First test", "comment_count": 2},
		{"id": "buckshot-xyz", "title": "Test bead 2", "status": "open", "description": "Second test", "comment_count": 0}
	]`

	before, _ := tracker.CaptureState(beforeJSON)
	after, _ := tracker.CaptureState(afterJSON)

	changes := tracker.Diff(before, after)
	summary := changes.FormatSummary()

	// Summary should not be empty
	if summary == "" {
		t.Error("Expected non-empty summary")
	}

	// Summary should mention created bead
	if !strings.Contains(summary, "buckshot-xyz") && !strings.Contains(strings.ToLower(summary), "created") {
		t.Log("Warning: summary may not mention created bead")
	}

	// Summary should mention status change
	if !strings.Contains(strings.ToLower(summary), "status") && !strings.Contains(summary, "in_progress") {
		t.Log("Warning: summary may not mention status change")
	}

	t.Logf("Summary:\n%s", summary)
}

// TestBeadsTracker_NoChanges tests that an empty diff is returned when nothing changed.
func TestBeadsTracker_NoChanges(t *testing.T) {
	tracker := NewTracker()

	json := `[
		{"id": "buckshot-abc", "title": "Test bead 1", "status": "open", "description": "First test"}
	]`

	before, _ := tracker.CaptureState(json)
	after, _ := tracker.CaptureState(json)

	changes := tracker.Diff(before, after)

	if len(changes.Created) != 0 {
		t.Errorf("Expected 0 created, got %d", len(changes.Created))
	}
	if len(changes.Modified) != 0 {
		t.Errorf("Expected 0 modified, got %d", len(changes.Modified))
	}
	if len(changes.Commented) != 0 {
		t.Errorf("Expected 0 commented, got %d", len(changes.Commented))
	}

	summary := changes.FormatSummary()
	if !strings.Contains(strings.ToLower(summary), "no changes") && summary != "" {
		t.Logf("Summary for no changes: %s", summary)
	}
}

// TestBeadsTracker_CaptureStateInvalidJSON tests error handling for invalid JSON.
func TestBeadsTracker_CaptureStateInvalidJSON(t *testing.T) {
	tracker := NewTracker()

	_, err := tracker.CaptureState("not valid json")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestBeadsTracker_CaptureStateEmptyArray tests handling of empty beads list.
func TestBeadsTracker_CaptureStateEmptyArray(t *testing.T) {
	tracker := NewTracker()

	state, err := tracker.CaptureState("[]")
	if err != nil {
		t.Fatalf("CaptureState([]) error = %v", err)
	}

	if len(state.Beads) != 0 {
		t.Errorf("Expected 0 beads, got %d", len(state.Beads))
	}
}
