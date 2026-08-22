// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

// Package tsicsparser provides functions for parsing ICS files.
import (
	"fmt"     // For formatting error messages.
	"strings" // For splitting strings.
	"time"    // For time operations.

	"github.com/thorsphere/tserr"   // For error handling.
	"github.com/thorsphere/tstable" // For creating table-like string representations of events.
)

// Event represents a calendar event with a summary and start time.
type Event struct {
	Uid      string    // Uid is the unique identifier for the event.
	Summary  string    // Summary is the title or description of the event.
	Location string    // Location is the location where the event takes place.
	Start    time.Time // Start is the start time of the event in UTC.
	End      time.Time // End is the end time of the event in UTC.
	// RRule holds the raw RFC 5545 §3.3.10 recurrence rule (e.g.
	// "FREQ=WEEKLY;BYDAY=MO,WE") if the event is recurring, or "" if it
	// is a single occurrence.
	//
	// IMPORTANT: the rule is stored verbatim and NOT expanded. Start and
	// End describe only the FIRST occurrence (DTSTART). Callers must not
	// treat a recurring event as a one-off; recurrence expansion is the
	// caller's responsibility.
	RRule string
}

// parseEvent parses an ICS event from the given scanner and returns an Event and an error if any.
func parseEvent(scanner *ICSScanner, tzs Timezones) (Event, error) {
	// Initialize an Event struct to hold the parsed event information.
	var event Event
	// pendingDuration holds a DURATION seen before DTSTART. A value-plus-flag
	// pair is used instead of *time.Duration to avoid a heap allocation:
	// time.Duration is a value type, and "set or not" semantics for a scalar
	// are expressed idiomatically in Go with an accompanying bool.
	// hasDuration is the authoritative "is it set?" indicator;
	// pendingDuration itself is only meaningful when hasDuration is true.
	var (
		pendingDuration time.Duration
		hasDuration     bool
		endSetByEnd     bool // true if event.End was set by DTEND (not DURATION)
	)
	// Loop through the lines of the scanner until END:VEVENT is encountered.
	for scanner.Scan() {
		// Read the next line from the scanner.
		line := scanner.Text()
		// Skip empty lines to avoid unnecessary processing.
		if len(line) == 0 {
			continue
		}
		// Split the line into key and value parts using the splitKeyValue function.
		parts, err := splitKeyValue(line)
		// If there is an error splitting the line, return an error indicating the issue.
		if err != nil {
			return Event{}, err
		}
		// Use a switch statement to handle different keys in the VEVENT component.
		switch parts.Key {
		// Handle the SUMMARY property, which specifies the event title.
		case "SUMMARY":
			event.Summary = parts.Value
		// Handle the UID property, which specifies the unique identifier for the event.
		case "UID":
			event.Uid = parts.Value
		// Handle the LOCATION property, which specifies the event location.
		case "LOCATION":
			event.Location = parts.Value
		// Handle the RRULE property: the raw recurrence rule. Stored verbatim;
		// expansion is deliberately out of scope for this version.
		case "RRULE":
			// RRULE may appear at most once per VEVENT.
			if event.RRule != "" {
				return Event{}, tserr.InvalidFormat("RRULE already set")
			}
			if err := validateRRule(parts.Value); err != nil {
				return Event{}, err
			}
			event.RRule = parts.Value
		// Handle the BEGIN of a nested component.
		case "BEGIN":
			// A nested sub-component inside VEVENT — most commonly
			// VALARM (RFC 5545 §3.6.6), but this arm handles any. This
			// package does not model alarms, so consume the entire block
			// (BEGIN through its matching END) and discard it. Without
			// this, the VALARM's body lines would be misparsed as
			// event-level properties and its END:VALARM would hit the
			// END switch's default arm, producing a spurious
			// "unexpected END:VALARM inside VEVENT" error. collectRawBlock
			// propagates structural errors (mismatched END, EOF) so a
			// malformed nested block still surfaces.
			if _, err := collectRawBlock(scanner, parts.Value); err != nil {
				return Event{}, err
			}
		// Handle the END of the VEVENT component.
		case "END":
			switch parts.Value {
			case "VEVENT": // If END:VEVENT is encountered, validate the event and return it.
				// --- Validation ---
				// DTSTART is REQUIRED per RFC 5545 §3.6.1.
				// If DTEND or DURATION was seen but DTSTART was not, that's an error.
				if event.Start.IsZero() {
					if !event.End.IsZero() {
						return Event{}, tserr.InvalidFormat("DTEND present without DTSTART in VEVENT")
					}
					if hasDuration {
						return Event{}, tserr.InvalidFormat("DURATION present without DTSTART in VEVENT")
					}
					// No DTSTART at all — also an error.
					return Event{}, tserr.InvalidFormat("missing required DTSTART in VEVENT")
				}
				// If only DTSTART was given (no DTEND, no DURATION), End defaults to Start.
				// This is valid per RFC 5545: a point-in-time event.
				if event.End.IsZero() {
					event.End = event.Start
				}
				// Successfully parsed the VEVENT component; return the event.
				return event, nil
			default: // If END is for any other component, return an error indicating unexpected END.
				return Event{}, tserr.InvalidFormat(
					fmt.Sprintf("unexpected END:%s inside VEVENT", parts.Value))
			}
		default:
			// Split the key into its base name and parameters (e.g., "DTSTART;TZID=US-Eastern").
			keyBase, params, err := splitKeyParams(parts.Key)
			// If there is an error splitting the key, return the error.
			if err != nil {
				return Event{}, err
			}
			switch keyBase {
			// If the key base is "DTSTART", parse the datetime and timezone ID.
			case "DTSTART":
				// DTSTART can only be used once per event. If it is already set, return an error.
				if !event.Start.IsZero() {
					return Event{}, tserr.InvalidFormat("DTSTART already set")
				}
				// Parse the DTSTART value into a UTC time.Time, handling both UTC-suffixed and TZID-qualified local times.
				if event.Start, err = parseDTValue(parts.Value, params, tzs, "DTSTART"); err != nil {
					return Event{}, err
				}
				// If a DURATION was seen before DTSTART, compute End now that Start is known.
				if hasDuration {
					event.End = event.Start.Add(pendingDuration)
					hasDuration = false
				}
			case "DTEND":
				// DTEND can only be used once per event. This is a dedicated
				// duplicate check, distinct from the DURATION mutual-exclusivity
				// check below, so the error message identifies the actual problem.
				// Note: event.End may already be set by a prior DTEND (caught here)
				// or by a DURATION (caught by the hasDuration check below). We
				// distinguish the two cases to give an actionable error.
				if endSetByEnd {
					return Event{}, tserr.InvalidFormat("DTEND already set; duplicate DTEND in VEVENT")
				}
				// DTEND and DURATION are mutually exclusive per RFC 5545 §3.6.1.
				// A pending DURATION (seen before DTSTART) counts as "DURATION set".
				if !event.End.IsZero() || hasDuration {
					return Event{}, tserr.InvalidFormat("DTEND and DURATION are mutually exclusive")
				}
				// Parse the DTEND value into a UTC time.Time, handling both UTC-suffixed and TZID-qualified local times.
				if event.End, err = parseDTValue(parts.Value, params, tzs, "DTEND"); err != nil {
					return Event{}, err
				}
				endSetByEnd = true
			case "DURATION":
				// DTEND and DURATION can only be used once per event as well as DTEND and DURATION are mutually exclusive.
				// If it is already set, return an error.
				if !event.End.IsZero() {
					return Event{}, tserr.InvalidFormat("DTEND already set or DURATION already set; DTEND and DURATION are mutually exclusive")
				}
				// DURATION already set
				if hasDuration {
					return Event{}, tserr.InvalidFormat("DURATION already set")
				}
				// Parse the duration string into a time.Duration.
				dur, err := parseDuration(parts.Value)
				// If there is an error parsing the duration, return the error.
				if err != nil {
					return Event{}, err
				}
				// If DTSTART has already been parsed, compute End immediately.
				if !event.Start.IsZero() {
					// Set the End time by adding the duration to the Start time.
					event.End = event.Start.Add(dur)
				} else {
					// DTSTART hasn't been seen yet. Store the duration temporarily.
					pendingDuration = dur
					hasDuration = true
				}
			}
		}
	}
	// Check for any scanning errors that may have occurred during the parsing process.
	if err := scanner.Err(); err != nil {
		return Event{}, err
	}
	// END:VEVENT was never encountered.
	// Return an error indicating that the VEVENT component was not properly closed.
	return Event{}, tserr.NotFound("END:VEVENT")
}

// parseDTValue parses a DTSTART or DTEND property value into a UTC time.Time.
// It handles both UTC-suffixed datetimes (e.g. "20250101T120000Z") and
// TZID-qualified local times (e.g. with params["TZID"] set), converting the
// latter to UTC via the calendar's timezone rules.
//
// propName ("DTSTART" or "DTEND") is used only to construct error messages
// when the value is a floating time (no TZID and no Zulu suffix).
func parseDTValue(value string, params map[string]string, tzs Timezones, propName string) (time.Time, error) {
	// Parse the local datetime from the value.
	localTime, isUTC, err := parseICSDateTime(value)
	if err != nil {
		return time.Time{}, err
	}
	// If the datetime is already UTC (Zulu suffix), no conversion is needed.
	if isUTC {
		return localTime, nil
	}
	// Not UTC — a TZID parameter is required to convert via the timezone rules.
	tzid, ok := params["TZID"]
	if !ok {
		return time.Time{}, tserr.InvalidFormat(fmt.Sprintf(
			"floating time not supported: %s without TZID or Zulu suffix; "+
				"floating times cannot be converted to UTC as required by this package",
			propName))
	}
	// Look up the timezone by its TZID in the provided Timezones map.
	tz, err := tzs.Lookup(tzid)
	// If the timezone with the given TZID is not found, return an error.
	if err != nil {
		return time.Time{}, err
	}
	// Convert the local time to UTC using the timezone rules.
	return convertToUTC(localTime, tzid, tz)
}

// splitKeyParams splits a key string into its base name and parameters.
// The function returns the base name, a map of parameters, and an error if any.
//
// Per RFC 5545 §3.2, parameter values may be quoted with double quotes,
// e.g. DTSTART;TZID="US-Eastern":20250101T120000. Quoting is used when the
// value contains characters that would otherwise be ambiguous (colons,
// semicolons, commas). The quotes are not part of the value and are stripped
// here so that downstream comparisons (e.g. tzid != tz.TZID in convertToUTC)
// compare the bare value.
//
// Only a single surrounding pair of double quotes is stripped; embedded or
// unbalanced quotes are left intact, matching the conservative behavior of
// most ICS parsers. A parameter value that is just a pair of quotes (e.g.
// TZID="") yields an empty string, which is a valid (if unusual) value.
func splitKeyParams(key string) (string, map[string]string, error) {
	// Split the key into its base name and parameters using the semicolon as a delimiter.
	parts := strings.Split(key, ";")
	// The first part is the base key (e.g., "DTSTART").
	keyBase := parts[0]
	// Initialize a map to hold the parameters (e.g., "TZID=US-Eastern").
	params := make(map[string]string)
	// Loop through the remaining parts to extract parameters.
	for _, param := range parts[1:] {
		// Split each parameter into key and value using the equals sign as a delimiter.
		paramParts := strings.SplitN(param, "=", 2)
		// If there are not exactly two parts (key and value), return an error.
		if len(paramParts) != 2 {
			return keyBase, nil, tserr.InvalidFormat(fmt.Sprintf("invalid parameter format: %s", param))
		}
		// Strip an optional surrounding pair of double quotes from the value.
		// RFC 5545 §3.2 allows parameter values to be quoted, e.g.
		// TZID="US-Eastern". The quotes are syntactic delimiters, not part of
		// the value, so they must be removed before storing the value;
		// otherwise downstream comparisons (tzid != tz.TZID) would fail on
		// quoted input. Only a single balanced outer pair is stripped —
		// embedded quotes or unbalanced quotes are left as-is.
		value := paramParts[1]
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
		params[paramParts[0]] = value
	}
	// Return the base key and the map of parameters.
	return keyBase, params, nil
}

// String returns a formatted string representation of the Event struct.
// It uses the tstable package to create a table-like output.
func (ev Event) String() string {
	// Create a new table with the event's UID as the title row
	tbl, err := tstable.New([]string{"UID", ev.Uid})
	// If there is an error creating the table, return an error string
	if err != nil {
		return fmt.Sprintf("<error creating table: %v>", err)
	}
	// Add rows for each field of the event
	tbl.AddRow([]string{"Summary", ev.Summary})
	tbl.AddRow([]string{"Location", ev.Location})
	tbl.AddRow([]string{"Start", ev.Start.Format(time.RFC3339)})
	tbl.AddRow([]string{"End", ev.End.Format(time.RFC3339)})
	// If the event has an RRule, add its value to the table
	if ev.RRule != "" {
		tbl.AddRow([]string{"RRule", ev.RRule})
	}
	// Set the multi-line flag for the second column (UID)
	tbl.SetMultiline(ev.Uid)
	// Return the formatted table as a string
	return tbl.String()
}

// validateRRule performs cheap syntactic validation of an RRULE value
// without interpreting it: FREQ is required and must be one of the seven
// defined values, and UNTIL and COUNT MUST NOT both be present (RFC 5545
// §3.3.10). Expansion semantics (BY* rules, INTERVAL, ...) are not
// validated — the rule is stored verbatim for the caller.
func validateRRule(s string) error {
	// Return an error if the RRULE string is empty, as it cannot be parsed.
	if s == "" {
		return tserr.InvalidFormat("empty RRULE")
	}
	// Initialize flags to indicate whether FREQ, UNTIL, and COUNT are present.
	var hasFreq, hasUntil, hasCount bool
	// Iterate over each part of the RRULE string.
	for part := range strings.SplitSeq(s, ";") {
		// Split each part by the first "=" to separate the key and value.
		kv := strings.SplitN(part, "=", 2)
		// If the split does not result in exactly two parts (key and value),
		// return an error indicating invalid format.
		if len(kv) != 2 {
			return tserr.InvalidFormat(fmt.Sprintf("invalid RRULE part: %s", part))
		}
		// Use a switch statement to handle different keys in the RRULE.
		switch kv[0] {
		case "FREQ": // FREQ is required
			switch kv[1] { // If the FREQ value is one of the seven defined values, set hasFreq to true.
			case "SECONDLY", "MINUTELY", "HOURLY", "DAILY", "WEEKLY", "MONTHLY", "YEARLY":
				hasFreq = true
			default: // If the FREQ value is not one of the seven defined values, return an error.
				return tserr.InvalidFormat(fmt.Sprintf("invalid RRULE FREQ value: %s", kv[1]))
			}
		case "UNTIL": // UNTIL is optional
			// If the UNTIL value is not a valid datetime, return an error.
			if _, _, err := parseICSDateTime(kv[1]); err != nil {
				return tserr.InvalidFormat(fmt.Sprintf("invalid RRULE UNTIL value: %s", kv[1]))
			}
			hasUntil = true
		case "COUNT": // COUNT is optional
			hasCount = true
		}
	}
	// If FREQ is not present, return an error.
	if !hasFreq {
		return tserr.InvalidFormat("RRULE missing required FREQ")
	}
	// If UNTIL and COUNT are both present, return an error.
	if hasUntil && hasCount {
		return tserr.InvalidFormat("RRULE must not contain both UNTIL and COUNT")
	}
	// TODO: validate BY* rules
	// TODO: validate INTERVAL
	// TODO: validate WKST
	// TODO: validate COUNT
	// TODO: validate BYSETPOS
	// TODO: validate BYDAY
	// TODO: validate BYMONTHDAY
	// TODO: validate BYYEARDAY
	// TODO: validate BYWEEKNO
	// TODO: validate BYMONTH
	// TODO: RFC 5545 §3.3.10 allows UNTIL to be a date-only value (UNTIL=20250401)
	// when DTSTART is a date — but parseICSDateTime only accepts datetime layouts,
	// so a date-only UNTIL is rejected here.
	// Return nil if all checks pass to indicate a valid RRULE.
	return nil
}
