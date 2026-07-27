// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

// Package tsicsparser provides functions for parsing ICS files.
import (
	"fmt"     // For formatting error messages.
	"strconv" // For parsing integers.
	"strings" // For splitting strings.
	"time"    // For time operations.

	"github.com/thorsphere/tserr" // For error handling.
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
			// If END:VEVENT, return the parsed event successfully.
			case "VEVENT":
				return event, nil
			// If END is for any other component, return an error indicating unexpected END.
			default:
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
			// If the key base is "DTSTART", parse the datetime and timezone ID.
			if keyBase == "DTSTART" {
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

// convertToUTC converts a local time to UTC using the given timezone.
// It returns the converted time and an error if any. If the TZID is empty,
// it returns an error. If the TZID does not match the timezone's TZID,
// it returns an error. If the timezone has no rules, it returns an error.
func convertToUTC(localTime time.Time, tzid string, tz Timezone) (time.Time, error) {
	// Return an error if the TZID is empty.
	if tzid == "" {
		return localTime, tserr.Empty("TZID")
	}
	// Return an error if the TZID does not match the timezone's TZID.
	if tzid != tz.TZID {
		return localTime, tserr.InvalidFormat(fmt.Sprintf("TZID mismatch: %s != %s", tzid, tz.TZID))
	}
	// Return an error if the timezone has no rules.
	if len(tz.Rules) == 0 {
		return localTime, tserr.InvalidFormat("no rules in timezone")
	}
	// Initialize pointers to Daylight and Standard rules.
	var dr, sr *TimezoneRule
	// Loop through the rules in the timezone to find Daylight and Standard rules.
	for _, rule := range tz.Rules {
		switch rule.Type {
		case Daylight:
			dr = &rule
		case Standard:
			sr = &rule
		}
	}
	// If both Daylight and Standard rules do not exist, return an error.
	if dr == nil && sr == nil {
		return localTime, tserr.InvalidFormat("no daylight and no standard rule in timezone")
	} else if dr == nil {
		// Only a Standard rule exists, no DST transitions.
		// Apply the Standard rule's TZOffsetTo directly to convert local time to UTC.
		return localTime.Add(-time.Duration(sr.TZOffsetTo) * time.Second), nil
	} else if sr == nil {
		// Only a Daylight rule exists, no Standard transitions.
		// Apply the Daylight rule's TZOffsetTo directly to convert local time to UTC.
		return localTime.Add(-time.Duration(dr.TZOffsetTo) * time.Second), nil
	}
	// Both Daylight and Standard rules exist.
	// Compute transition times for both rules in the year of localTime.
	year := localTime.Year()
	// Compute the transition time for the Daylight rule in the year.
	daylight, err := transitionTime(*dr, year)
	if err != nil {
		return localTime, err
	}
	// Compute the transition time for the Standard rule in the year.
	standard, err := transitionTime(*sr, year)
	if err != nil {
		return localTime, err
	}
	// Determine which transition happens first in the year.
	// After a transition, that rule's TZOffsetTo is in effect.
	var offset int
	if daylight.Before(standard) {
		// Northern Hemisphere: spring forward (daylight), fall back (standard).
		// Before daylight transition or after standard transition → Standard.
		// Between daylight and standard → Daylight.
		if localTime.Before(daylight) || !localTime.Before(standard) {
			offset = sr.TZOffsetTo
		} else {
			offset = dr.TZOffsetTo
		}
	} else {
		// Southern Hemisphere: spring back (standard), fall forward (daylight).
		// Before standard transition or after daylight transition → Daylight.
		// Between standard and daylight → Standard.
		if localTime.Before(standard) || !localTime.Before(daylight) {
			offset = dr.TZOffsetTo
		} else {
			offset = sr.TZOffsetTo
		}
	}
	// Apply the offset to convert local time to UTC.
	return localTime.Add(-time.Duration(offset) * time.Second), nil
}

// transitionTime computes the local datetime when a timezone rule takes effect
// in the given year. It uses the rule's RRULE (if present) to determine the
// recurring date, falling back to the rule's DTSTART for fixed transitions.
func transitionTime(rule TimezoneRule, year int) (time.Time, error) {
	// If the rule has no recurrence rule, the transition is a fixed one-time event.
	// Return the DTSTART as-is (only meaningful for the year it was defined in).
	if rule.RRule == nil {
		return rule.DTStart, nil
	}

	// Determine the month from the RRULE's ByMonth field.
	month := time.Month(rule.RRule.ByMonth)
	// Compute the day-of-month for the Nth occurrence of the weekday in that month.
	day, err := nthWeekday(year, month, rule.RRule.ByDay)

	if err != nil {
		return time.Time{}, err
	}

	// Construct the transition datetime using the time-of-day from DTStart.
	return time.Date(year, month, day,
		rule.DTStart.Hour(), rule.DTStart.Minute(), rule.DTStart.Second(), 0, time.UTC), nil
}

// nthWeekday returns the day-of-month for the Nth occurrence of a weekday
// in the given month and year. The byDay string is in RRULE BYDAY format,
// e.g., "2SU" (2nd Sunday), "-1SU" (last Sunday), "1FR" (1st Friday).
// If the BYDAY value is only two characters (e.g., "SU"), there is no
// ordinal prefix and the function returns 0 for ordinal (treating the
// Nth occurrence as the first, matching RFC 5545 semantics).
func nthWeekday(year int, month time.Month, byDay string) (int, error) {
	// Parse the ordinal and weekday abbreviation from the BYDAY value.
	ordinal, wd, err := parseByDay(byDay)
	// If there is an error parsing the BYDAY value, return zero and the error.
	if err != nil {
		return 0, err
	}
	// Positive ordinal: count forward from the start of the month.
	if ordinal > 0 {
		// Get the first day of the month and its weekday.
		first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		firstWeekday := first.Weekday()
		// Calculate the offset from the 1st to the first occurrence of the target weekday.
		offset := (int(wd) - int(firstWeekday) + 7) % 7
		// Return the day-of-month for the Nth occurrence.
		return 1 + offset + (ordinal-1)*7, nil
	}
	// Negative ordinal: count backward from the end of the month.
	if ordinal < 0 {
		// Get the last day of the month and its weekday.
		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		lastWeekday := time.Date(year, month, lastDay, 0, 0, 0, 0, time.UTC).Weekday()

		// Calculate the offset from the last day back to the last occurrence of the target weekday.
		backOffset := (int(lastWeekday) - int(wd) + 7) % 7
		lastOccurrence := lastDay - backOffset

		// Move back by (|ordinal| - 1) weeks. Since ordinal is negative, (ordinal+1)*7
		// gives the correct offset (e.g., -1 → 0 weeks back, -2 → -7 days back).
		return lastOccurrence + (ordinal+1)*7, nil
	}
	// If the ordinal is not positive or negative, return an error.
	return 0, tserr.InvalidFormat(fmt.Sprintf("invalid ordinal: %d", ordinal))
}

// parseByDay parses an RRULE BYDAY value like "2SU" or "-1SU" into
// an ordinal (positive or negative) and a time.Weekday.
// If no ordinal is present (e.g., "SU"), it defaults to 1.
func parseByDay(byDay string) (int, time.Weekday, error) {
	// The last two characters are always the weekday abbreviation.
	wa := byDay[len(byDay)-2:]
	// Parse the weekday abbreviation.
	wd, err := parseWeekday(wa)
	// If there is an error parsing the weekday, return zero and the error.
	if err != nil {
		return 0, -1, err
	}
	// If the BYDAY value is only two characters, there is no ordinal prefix.
	if len(byDay) == 2 {
		return 0, wd, nil
	}
	// Extract the ordinal prefix (e.g., "2" from "2SU", "-1" from "-1SU").
	op := byDay[:len(byDay)-2]
	// Parse the ordinal prefix (e.g., "2" from "2SU", "-1" from "-1SU").
	ordinal, err := strconv.Atoi(op)
	// If there is an error parsing the ordinal, return zero and the error.
	if err != nil {
		// If parsing fails, return zero, the weekday and the error.
		return 0, wd, tserr.InvalidFormat(fmt.Sprintf("invalid ordinal: %s", op))
	}
	// Return the parsed ordinal and weekday.
	return ordinal, wd, nil
}

// parseWeekday parses a string representing a weekday and returns the corresponding time.Weekday.
// It returns an error if the weekday is not recognized.
func parseWeekday(s string) (time.Weekday, error) {
	// Switch on the weekday string to determine the corresponding time.Weekday.
	switch s {
	case "SU":
		return time.Sunday, nil
	case "MO":
		return time.Monday, nil
	case "TU":
		return time.Tuesday, nil
	case "WE":
		return time.Wednesday, nil
	case "TH":
		return time.Thursday, nil
	case "FR":
		return time.Friday, nil
	case "SA":
		return time.Saturday, nil
	default: // If the weekday is not recognized, return an error.
		return -1, tserr.InvalidFormat(fmt.Sprintf("invalid weekday: %s", s))
	}
}

// splitKeyParams splits a key string into its base name and parameters.
// The function returns the base name, a map of parameters, and an error if any.
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
