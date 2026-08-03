// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

// Import the necessary packages for formatting, mathematical operations, string conversion,
// time handling, and custom error handling.
import (
	"fmt"     // For formatting error messages
	"math"    // For mathematical operations, specifically to get the maximum int64 value
	"strconv" // For converting strings to integers
	"time"    // For handling time durations

	"github.com/thorsphere/tserr" // Custom error handling package for the project
)

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
