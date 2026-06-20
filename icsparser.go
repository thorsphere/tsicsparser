// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsicsparser

import "time"

// Calendar represents a calendar with events.
type Calendar struct {
	Events []Event
}

// Event represents a calendar event with a summary and start time.
type Event struct {
	Summary  string
	Start    time.Time
}
