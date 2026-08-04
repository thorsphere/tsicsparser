// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

// Import necessary packages for time operations, string handling, and error management.
import (
	"fmt"     // For formatting error messages.
	"strconv" // For parsing integers.
	"strings" // For splitting strings.
	"time"    // For time operations.

	"github.com/thorsphere/tserr" // For error handling.
)

// convertToUTC converts a local time to UTC using the given timezone.
// It returns the converted time and an error if any. If the TZID is empty,
// it returns an error. If the TZID does not match the timezone's TZID,
// it returns an error. If the timezone has no rules, it returns an error.
//
// Offset selection convention:
//
//   - Only TZOffsetTo (the offset in effect AFTER a transition) is used to
//     compute UTC. TZOffsetFrom (the offset BEFORE the transition) is not
//     consulted; it is assumed to be the TZOffsetTo of the opposing rule.
//     This package does not validate that TZOffsetFrom of one rule equals
//     the TZOffsetTo of the other — malformed timezones where this
//     invariant does not hold will produce incorrect (but deterministic)
//     conversions.
//
//   - At a transition instant itself, the NEW offset (the one taking effect
//     at that instant) is applied. That is, localTime == daylightTransition
//     yields the daylight offset, and localTime == standardTransition
//     yields the standard offset. This matches the common "the transition
//     has just occurred" interpretation and is consistent across both
//     hemispheres and both transition points.
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
	for i := range tz.Rules {
		rule := &tz.Rules[i]
		switch rule.Type {
		case Daylight:
			dr = rule
		case Standard:
			sr = rule
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
	// Validate that each rule's TZOffsetFrom matches the other
	// rule's TZOffsetTo. This catches malformed VTIMEZONE blocks early.
	if dr.TZOffsetFrom != sr.TZOffsetTo {
		return localTime, tserr.InvalidFormat(fmt.Sprintf("DAYLIGHT TZOffsetFrom (%d) != STANDARD TZOffsetTo (%d)", dr.TZOffsetFrom, sr.TZOffsetTo))
	}
	if sr.TZOffsetFrom != dr.TZOffsetTo {
		return localTime, tserr.InvalidFormat(fmt.Sprintf("STANDARD TZOffsetFrom (%d) != DAYLIGHT TZOffsetTo (%d)", sr.TZOffsetFrom, dr.TZOffsetTo))
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
	// --- Unified offset selection (hemisphere-agnostic) ---
	//
	// Both hemispheres follow the same rule once the transitions are ordered:
	//
	//   - first  = the earlier transition in the year, with its rule (firstRule)
	//   - second = the later  transition in the year, with its rule (secondRule)
	//
	//   Period              | Offset in effect
	//   --------------------|---------------------------
	//   before first        | secondRule.TZOffsetTo  (outside)
	//   [first, second)     | firstRule.TZOffsetTo   (between)
	//   at or after second  | secondRule.TZOffsetTo  (outside)
	//
	// Why this is correct for both hemispheres:
	//
	//   Northern: daylight < standard.
	//     first=daylight (dr), second=standard (sr).
	//     between  → dr.TZOffsetTo (Daylight)
	//     outside → sr.TZOffsetTo (Standard)
	//
	//   Southern: standard < daylight.
	//     first=standard (sr), second=daylight (dr).
	//     between  → sr.TZOffsetTo (Standard)
	//     outside → dr.TZOffsetTo (Daylight)
	//
	// Boundary convention (unchanged): at a transition instant, the NEW
	// offset is selected. localTime == first falls into the "between" branch
	// (firstRule's offset, which is the one taking effect at 'first');
	// localTime == second falls into the "outside" branch (secondRule's
	// offset, which is the one taking effect at 'second').
	var first, second time.Time
	var firstRule, secondRule *TimezoneRule
	if daylight.Before(standard) {
		first, firstRule = daylight, dr
		second, secondRule = standard, sr
	} else {
		first, firstRule = standard, sr
		second, secondRule = daylight, dr
	}

	var offset int
	switch {
	case localTime.Before(first):
		// Before the first transition: the second transition's offset is
		// still in effect from the previous cycle.
		offset = secondRule.TZOffsetTo
	case !localTime.Before(second):
		// At or after the second transition: the second transition's offset
		// is in effect for the rest of the year.
		offset = secondRule.TZOffsetTo
	default:
		// Between the two transitions: the first transition's offset is
		// in effect.
		offset = firstRule.TZOffsetTo
	}

	// Apply the offset to convert local time to UTC.
	return localTime.Add(-time.Duration(offset) * time.Second), nil
}

// transitionTime computes the local datetime when a timezone rule takes effect
// in the given year. It uses the rule's RRULE (if present) to determine the
// recurring date, falling back to the rule's DTSTART for fixed transitions.
//
// Fixed (non-recurring) rules:
//
//   - Per RFC 5545, a sub-component without RRULE defines a one-time
//     transition at its DTSTART. However, this package uses transitionTime
//     to convert local times that may fall in any year. Returning DTSTART
//     verbatim for a different year would yield a stale transition date and
//     produce incorrect offset selection in convertToUTC.
//
//   - To handle this, a fixed rule's DTSTART is treated as a template: the
//     month, day-of-month, and time-of-day are reused, but the year is
//     replaced with the requested 'year'. This is correct for fixed-date
//     transitions (e.g. "always January 1 at 00:00") and is a reasonable
//     approximation for legacy rules that omit RRULE but intend annual
//     recurrence.
//
//   - If the requested year differs from DTSTART's year and the rule is
//     genuinely one-time (not annual), the reconstructed date is still a
//     best-effort approximation. Callers requiring strict one-time
//     semantics should validate rule.DTStart.Year() == year before relying
//     on the result.
func transitionTime(rule TimezoneRule, year int) (time.Time, error) {
	// If the rule has no recurrence rule, the transition is a fixed one.
	// Reconstruct it for the requested year using DTStart's month, day,
	// and time-of-day as a template, so that convertToUTC compares against
	// a transition date in the same year as localTime.
	if rule.RRule == nil {
		// Fast path: the requested year matches DTStart's year, so return
		// DTStart verbatim. This preserves exact behavior for the common
		// case and avoids any rounding from the Date reconstruction.
		if year == rule.DTStart.Year() {
			return rule.DTStart, nil
		}
		// Reconstruct the transition in the requested year. time.Date
		// normalizes out-of-range values, so if the template day-of-month
		// is invalid for the target year (e.g. Feb 29 in a non-leap year),
		// Go rolls it forward into March. We detect that by checking the
		// resulting month and return an error rather than silently using
		// a wrong date.
		result := time.Date(year,
			rule.DTStart.Month(),
			rule.DTStart.Day(),
			rule.DTStart.Hour(),
			rule.DTStart.Minute(),
			rule.DTStart.Second(),
			rule.DTStart.Nanosecond(),
			time.UTC)
		if result.Month() != rule.DTStart.Month() {
			return time.Time{}, tserr.InvalidFormat(fmt.Sprintf(
				"fixed transition %s does not exist in year %d (day %d of %s out of bounds)",
				rule.DTStart.Format("2006-01-02"), year,
				rule.DTStart.Day(), rule.DTStart.Month()))
		}
		return result, nil
	}
	// Determine the month from the RRULE's ByMonth field.
	month := time.Month(rule.RRule.ByMonth)
	// If the RRULE's BYMONTH is missing (0), fall back to the month from DTSTART.
	// This handles legacy or malformed rules where BYMONTH is absent.
	if month == 0 {
		month = rule.DTStart.Month()
	}
	// Compute the day-of-month for the Nth occurrence of the weekday in that month.
	day, err := nthWeekday(year, month, rule.RRule.ByDay)
	// If the Nth occurrence is out of bounds, return an error.
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
// If the BYDAY value has no ordinal prefix (e.g., "SU"), it returns an error
// because weekly recurrence is invalid within VTIMEZONE definitions.
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
		// Calculate the day-of-month for the Nth occurrence.
		targetDay := 1 + offset + (ordinal-1)*7
		// If the target day is outside the month, return an error.
		if time.Date(year, month, targetDay, 0, 0, 0, 0, time.UTC).Month() != month {
			return 0, tserr.InvalidFormat(fmt.Sprintf("occurrence %d of %s out of bounds for month %s", ordinal, byDay, month))
		}
		return targetDay, nil
	}
	// Negative ordinal: count backward from the end of the month.
	if ordinal < 0 {
		// Get the last day of the month and its weekday.
		last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC)
		lastDay := last.Day()
		lastWeekday := last.Weekday()
		// Calculate the offset from the last day back to the last occurrence of the target weekday.
		backOffset := (int(lastWeekday) - int(wd) + 7) % 7
		lastOccurrence := lastDay - backOffset
		// Move back by (|ordinal| - 1) weeks. Since ordinal is negative, (ordinal+1)*7
		// gives the correct offset (e.g., -1 → 0 weeks back, -2 → -7 days back).
		targetDay := lastOccurrence + (ordinal+1)*7
		// --- Bounds checks (symmetric with the positive branch) ---
		// The positive branch validates via time.Date(...).Month() != month,
		// which catches both underflow (day < 1) and overflow (day > lastDay).
		// Although the arithmetic makes overflow
		// impossible in practice (a negative ordinal only moves backward
		// from lastOccurrence, so targetDay <= lastOccurrence <= lastDay),
		// we add an explicit upper-bound assertion for symmetry and to
		// guard against future changes to the offset formula.
		//
		// Lower bound: the day must not fall before the 1st of the month.
		if targetDay < 1 {
			return 0, tserr.InvalidFormat(fmt.Sprintf("negative occurrence %d of %s out of bounds for month %s", ordinal, byDay, month))
		}
		// Upper bound: the day must not fall after the last day of the month.
		// This is the symmetric counterpart to the positive branch's
		// time.Date(...).Month() != month check.
		if targetDay > lastDay {
			return 0, tserr.InvalidFormat(fmt.Sprintf("negative occurrence %d of %s out of bounds for month %s (day %d > %d)", ordinal, byDay, month, targetDay, lastDay))
		}
		return targetDay, nil
	}
	// If ordinal is 0, no modifier was provided (e.g., "SU").
	// This means "every Sunday", which is invalid/unsupported inside a VTIMEZONE block.
	return 0, tserr.InvalidFormat(fmt.Sprintf("unsupported weekday rule without ordinal prefix inside VTIMEZONE: %s", byDay))
}

// parseByDay parses an RRULE BYDAY value like "2SU" or "-1SU" into
// an ordinal (positive or negative) and a time.Weekday.
// If no ordinal is present (e.g., "SU"), it returns 0 for the ordinal,
// which will be rejected as an error by the caller inside VTIMEZONE context.
//
// Error convention: on any error, both the ordinal and weekday are returned
// as their zero values (0 and time.Sunday respectively). Callers MUST check
// err first and disregard the returned values when err != nil; the zero
// values are placeholders, not meaningful results.
func parseByDay(byDay string) (int, time.Weekday, error) {
    // Safeguard against out-of-bounds access: Check whether the string is long enough.
    // Since a day of the week always consists of at least 2 characters (e.g., "SU"),
    // anything shorter than that is definitely invalid.
    if len(byDay) < 2 {
        return 0, 0, tserr.InvalidFormat(fmt.Sprintf("invalid BYDAY value: %s", byDay))
    }
    // The last two characters are always the weekday abbreviation.
    wa := byDay[len(byDay)-2:]
    // Parse the weekday abbreviation.
    wd, err := parseWeekday(wa)
    // If there is an error parsing the weekday, return zero values and the error.
    if err != nil {
        return 0, 0, err
    }
    // If the BYDAY value is only two characters, there is no ordinal prefix.
    if len(byDay) == 2 {
        return 0, wd, nil
    }
    // Extract the ordinal prefix (e.g., "2" from "2SU", "-1" from "-1SU").
    op := byDay[:len(byDay)-2]
    // Parse the ordinal prefix (e.g., "2" from "2SU", "-1" from "-1SU").
    ordinal, err := strconv.Atoi(op)
    // If there is an error parsing the ordinal, return zero for the ordinal,
    // the successfully-parsed weekday, and the error. The weekday is valid
    // here (it parsed above), so returning it is accurate; only the ordinal
    // failed. This lets a hypothetical caller that only needs the weekday
    // recover it, while callers needing the ordinal must check err.
    if err != nil {
        return 0, wd, tserr.InvalidFormat(fmt.Sprintf("invalid ordinal prefix: %s in BYDAY %s", op, byDay))
    }
    // Return the parsed ordinal and weekday.
    return ordinal, wd, nil
}

// parseWeekday parses a string representing a weekday and returns the corresponding time.Weekday.
// It returns an error if the weekday is not recognized.
func parseWeekday(s string) (time.Weekday, error) {
	// Switch on the weekday string to determine the corresponding time.Weekday.
	switch strings.ToUpper(s) {
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
