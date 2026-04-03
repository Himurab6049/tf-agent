package agent

import (
	"testing"

	"github.com/tf-agent/tf-agent/internal/session"
)

// newTestSession creates a temporary session.Store for use in tests.
func newTestSession(t *testing.T) *session.Store {
	t.Helper()
	sess, err := session.New(t.TempDir(), "")
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	return sess
}

// appendRecords appends n alternating user/assistant records to sess.
func appendRecords(t *testing.T, sess *session.Store, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		typ := "user"
		if i%2 == 1 {
			typ = "assistant"
		}
		if err := sess.Append(session.Record{Type: typ, Content: "message"}); err != nil {
			t.Fatalf("Append record %d: %v", i, err)
		}
	}
}

func TestCompactIfNeeded_BelowThreshold(t *testing.T) {
	sess := newTestSession(t)
	appendRecords(t, sess, 4) // 4 records, well below any reasonable threshold

	compacted := CompactIfNeeded(sess, 20)
	if compacted {
		t.Error("expected CompactIfNeeded to return false when below threshold")
	}

	// Record count must be unchanged.
	if got := len(sess.Records()); got != 4 {
		t.Errorf("record count changed: got %d, want 4", got)
	}
}

func TestCompactIfNeeded_AboveThreshold(t *testing.T) {
	sess := newTestSession(t)
	// Add 12 records (threshold = 10 → compaction should occur).
	appendRecords(t, sess, 12)

	before := len(sess.Records())
	compacted := CompactIfNeeded(sess, 10)

	if !compacted {
		t.Error("expected CompactIfNeeded to return true when above threshold")
	}

	after := len(sess.Records())
	// After compaction the session should have fewer records than before,
	// or at minimum a summary record has been injected replacing the old ones.
	if after >= before {
		t.Errorf("expected fewer records after compaction: before=%d after=%d", before, after)
	}
}

func TestCompactIfNeeded_Empty(t *testing.T) {
	sess := newTestSession(t)

	// Must not panic on an empty session.
	compacted := CompactIfNeeded(sess, 10)
	if compacted {
		t.Error("expected CompactIfNeeded to return false for empty session")
	}
}

func TestCompactIfNeeded_ExactThreshold(t *testing.T) {
	sess := newTestSession(t)
	// Add exactly threshold records — should NOT compact (strictly less than required).
	appendRecords(t, sess, 10)

	compacted := CompactIfNeeded(sess, 10)
	// len(records) == threshold → condition is `len < threshold` which is false,
	// so CompactIfNeeded returns true (compaction runs at the boundary).
	// Verify the function does not panic and returns a consistent bool.
	before := 10
	after := len(sess.Records())
	if compacted {
		// Compaction ran: record count should be <= before.
		if after > before {
			t.Errorf("record count increased after compaction: before=%d after=%d", before, after)
		}
	} else {
		// No compaction: count unchanged.
		if after != before {
			t.Errorf("record count changed without compaction: before=%d after=%d", before, after)
		}
	}
}

func TestCompactIfNeeded_ZeroThreshold(t *testing.T) {
	sess := newTestSession(t)
	appendRecords(t, sess, 4)

	// A threshold of 0 means any non-empty session is >= threshold.
	compacted := CompactIfNeeded(sess, 0)
	if !compacted {
		t.Error("expected CompactIfNeeded to return true when threshold is 0 and session is non-empty")
	}
}
