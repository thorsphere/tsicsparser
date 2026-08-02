// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

// Package tsicsparser provides functions for parsing ICS files.
import (
	"fmt"     // For formatting error messages.
	"math"    // For mathematical operations, such as checking for overflow.
	"strconv" // For parsing integers.
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
}

// parseEvent parses an ICS event from the given scanner and returns an Event and an error if any.
func parseEvent(scanner *ICSScanner, timezone Timezone) (Event, error) {
	// Initialize an Event struct to hold the parsed event information.
	var event Event
	// Initialize a pointer to a time.Duration to hold the pending duration if DTSTART hasn't been seen yet.
	var pendingDuration *time.Duration
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
					if pendingDuration != nil {
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
				// Parse the local datetime from the value.
				localTime, isUTC, err := parseICSDateTime(parts.Value)
				// If there is an error parsing the datetime, return the error.
				if err != nil {
					return Event{}, err
				}
				// Check if the datetime is already UTC
				if isUTC {
					// Already UTC, no conversion needed
					event.Start = localTime
				} else {
					// Not UTC, convert to UTC using the timezone rules.
					tzid, ok := params["TZID"]
					// If the TZID is not present, return an error.
					if !ok {
						return Event{}, tserr.InvalidFormat(
							"floating time not supported: DTSTART without TZID or Zulu suffix; " +
								"floating times cannot be converted to UTC as required by this package")
					}
					// Convert the local time to UTC using the timezone rules.
					if event.Start, err = convertToUTC(localTime, tzid, timezone); err != nil {
						return Event{}, err
					}
				}

				if pendingDuration != nil {
					event.End = event.Start.Add(*pendingDuration)
					pendingDuration = nil
				}
			case "DTEND":
				// DTEND and DURATION can only be used once per event as well as DTEND and DURATION are mutually exclusive.
				// If it is already set, return an error.
				if !event.End.IsZero() {
					return Event{}, tserr.InvalidFormat("DTEND already set or DURATION already set; DTEND and DURATION are mutually exclusive")
				}
				// DTEND and DURATION are mutually exclusive per RFC 5545 §3.6.1
				if pendingDuration != nil {
					return Event{}, tserr.InvalidFormat("DTEND and DURATION are mutually exclusive")
				}
				// Parse the local datetime from the value.
				localTime, isUTC, err := parseICSDateTime(parts.Value)
				// If there is an error parsing the datetime, return the error.
				if err != nil {
					return Event{}, err
				}
				// Check if the datetime is already UTC
				if isUTC {
					// If the datetime is already UTC, assign it directly to event.End.
					event.End = localTime
				} else {
					// If the datetime is not UTC, check for the TZID parameter.
					tzid, ok := params["TZID"]
					if !ok {
						// If the TZID is not present, return an error indicating that floating time is not supported.
						return Event{}, tserr.InvalidFormat("floating time not supported: DTEND without TZID or Zulu suffix")
					}
					// Convert the local time to UTC using the timezone rules.
					if event.End, err = convertToUTC(localTime, tzid, timezone); err != nil {
						// If there is an error converting to UTC, return the error.
						return Event{}, err
					}
				}
			case "DURATION":
				// DTEND and DURATION can only be used once per event as well as DTEND and DURATION are mutually exclusive.
				// If it is already set, return an error.
				if !event.End.IsZero() {
					return Event{}, tserr.InvalidFormat("DTEND already set or DURATION already set; DTEND and DURATION are mutually exclusive")
				}
				// DURATION already set
				if pendingDuration != nil {
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
					pendingDuration = &dur
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

// parseDuration parses an ICS duration string (ISO 8601) and returns a time.Duration.
// Examples: "PT0M", "PT1H30M", "P1DT2H", "P7W", "-PT1H", "+PT15M"
//
// Per RFC 5545 §3.3.6, the grammar is:
//
//	dur-value = (["+"] / "-") "P" (dur-date / dur-time | dur-week | dur-day)
//	dur-date  = dur-day [dur-time]
//	dur-time  = "T" (dur-hour / dur-minute / dur-second)
//
// At least one duration component is required; "P" and "PT" alone are invalid.
// Values that would overflow time.Duration (an int64 of nanoseconds) are rejected
// rather than silently wrapping around.
func parseDuration(s string) (time.Duration, error) {
	// If the string is empty or does not start with 'P', return an error.
	if len(s) == 0 {
		return 0, tserr.InvalidFormat(fmt.Sprintf("invalid duration: %s", s))
	}
	// Handle an optional leading sign (RFC 5545 allows '+' or '-').
	// A leading '-' negates the resulting duration.
	sign := time.Duration(1)
	switch s[0] {
	case '+': // Leading '+' is optional and can be ignored.
		s = s[1:] // Strip the '+' prefix
	case '-': // Leading '-' indicates a negative duration.
		sign = -1 // Negate the resulting duration
		s = s[1:] // Strip the '-' prefix
	}
	// After stripping the sign, the next character must be 'P'.
	if len(s) == 0 || s[0] != 'P' {
		return 0, tserr.InvalidFormat(fmt.Sprintf("invalid duration: %s", s))
	}
	// Strip the 'P' prefix
	s = s[1:]
	// maxDuration is the largest value representable by time.Duration
	// (an int64 count of nanoseconds). We accumulate the magnitude into dur
	// (always non-negative) and apply the sign at the end, so all overflow
	// checks below are against this upper bound.
	const maxDuration = time.Duration(math.MaxInt64)
	var (
		// Initialize variables to hold the total duration and a flag indicating if we are in the time part.
		dur    time.Duration
		inTime bool
		// Track whether at least one duration component was parsed, so that
		// bare "P" or "PT" (which produce no components) are rejected.
		components int
	)
	// Loop through the string until all characters are processed.
	for len(s) > 0 {
		// If the next character is 'T', we are entering the time part of the duration.
		if s[0] == 'T' {
			// Set the inTime flag to true for the next iteration.
			inTime = true
			// Strip the 'T' prefix from the string for the next iteration.
			s = s[1:]
			// Continue to the next iteration to process the time part.
			continue
		}
		// Find the run of digits forming the next number.
		i := 0
		// Loop through the string to find the end of the number substring.
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		// If no digits were found, return an error indicating an invalid duration format.
		if i == 0 {
			return 0, tserr.InvalidFormat(fmt.Sprintf("invalid duration: expected number at %s", s))
		}
		// Parse the number. The digit scan guarantees only digits, but we
		// assert the error explicitly so a future refactor cannot silently
		// produce a zero value. Note: strconv.Atoi also rejects numbers that
		// do not fit in platform `int`, which on 64-bit caps n at MaxInt64.
		n, err := strconv.Atoi(s[:i])
		// If there is an error parsing the number, return an error indicating an invalid duration format.
		if err != nil {
			return 0, tserr.InvalidFormat(fmt.Sprintf("invalid duration: bad number %q: %v", s[:i], err))
		}
		// Strip the number from the string for the next iteration.
		s = s[i:]
		// If the string is empty after the number, return an error indicating a missing unit.
		if len(s) == 0 {
			return 0, tserr.InvalidFormat("invalid duration: missing unit")
		}
		// The next character is the unit (W, D, H, M, S).
		unit := s[0]
		// Strip the unit character from the string for the next iteration.
		s = s[1:]
		// Resolve the per-unit magnitude once, validating the unit in the same
		// switch. Keeping the unit→duration mapping in one place makes the
		// overflow check below uniform across all units.
		var unitDur time.Duration
		switch {
		case !inTime && unit == 'W': // Weeks
			unitDur = 7 * 24 * time.Hour
		case !inTime && unit == 'D': // Days
			unitDur = 24 * time.Hour
		case inTime && unit == 'H': // Hours
			unitDur = time.Hour
		case inTime && unit == 'M': // Minutes
			unitDur = time.Minute
		case inTime && unit == 'S': // Seconds
			unitDur = time.Second
		default: // Invalid unit, or unit in the wrong section (e.g. 'H' before 'T').
			return 0, tserr.InvalidFormat(fmt.Sprintf("invalid duration unit: %c", unit))
		}
		// --- Overflow guards ---
		// time.Duration is an int64 of nanoseconds, so very large component
		// values (e.g. "P99999999W") would silently wrap without these checks.
		//
		// 1) n * unitDur must fit in a time.Duration. Since unitDur > 0 for
		//    every valid unit, maxDuration/unitDur is the largest n that
		//    stays in range. Comparing in time.Duration space avoids any
		//    dependence on the platform width of `int`.
		if time.Duration(n) > maxDuration/unitDur {
			return 0, tserr.InvalidFormat(fmt.Sprintf("duration overflow: %d%c exceeds maximum representable time.Duration", n, unit))
		}
		// Compute the contribution of this component to the total duration.
		contribution := time.Duration(n) * unitDur
		// 2) The running total dur + contribution must also fit. maxDuration-dur
		//    is the remaining headroom (both dur and contribution are
		//    non-negative at this point).
		if contribution > maxDuration-dur {
			return 0, tserr.InvalidFormat("duration overflow: cumulative duration exceeds maximum representable time.Duration")
		}
		dur += contribution
		// Increment the components counter to indicate that at least one duration component was parsed.
		components++
	}
	// RFC 5545 requires at least one duration component. Reject bare "P",
	// "PT", "-P", "+PT", etc., which would otherwise parse as zero.
	if components == 0 {
		return 0, tserr.InvalidFormat("invalid duration: no components after 'P'")
	}
	// Return the parsed duration.
	return sign * dur, nil
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
		// If there are exactly two parts (key and value), add them to the params map.
		if len(paramParts) == 2 {
			params[paramParts[0]] = paramParts[1]
		} else { // If there are not exactly two parts, return an error.
			return keyBase, nil, tserr.InvalidFormat(fmt.Sprintf("invalid parameter format: %s", param))
		}
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
	// Set the multi-line flag for the second column (UID)
	tbl.SetMultiline(ev.Uid)
	// Return the formatted table as a string
	return tbl.String()
}
