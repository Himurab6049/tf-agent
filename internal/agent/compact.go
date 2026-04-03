package agent

import (
	"github.com/tf-agent/tf-agent/internal/session"
)

// CompactIfNeeded compacts the session if the record count exceeds threshold.
func CompactIfNeeded(sess *session.Store, threshold int) bool {
	records := sess.Records()
	if len(records) < threshold {
		return false
	}

	compacted := session.Compact(records, 5)
	sess.Clear()
	for _, r := range compacted {
		_ = sess.Append(r)
	}
	return true
}
