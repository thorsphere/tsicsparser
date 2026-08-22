// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

// Import the tserr package to handle error reporting and formatting in the functions.
import (
	"fmt"     // For formatting error messages
	"strconv" // For converting string representations of numbers to integers
	"strings" // For string manipulation functions like Split and SplitN

	"github.com/thorsphere/tserr" // For error handling and reporting
)

// RRule represents a recurrence rule for timezone transitions.
type RRule struct {
	Freq    string // e.g., "YEARLY"
	ByMonth int    // 1-12
	ByDay   string // e.g., "1SU", "-1SU"
}

// parseRRule parses an RRULE value for timezone transitions like "FREQ=YEARLY;BYMONTH=3;BYDAY=2SU".
func parseRRule(s string) (*RRule, error) {
	// Return an error if the RRULE string is empty, as it cannot be parsed.
	if s == "" {
		return nil, tserr.InvalidFormat("empty RRULE")
	}
	// Initialize an empty RRule struct to hold the parsed values.
	var rrule RRule
	// Split the RRULE string by semicolons to separate each key-value pair.
	// Iterate over each part of the RRULE string.
	for part := range strings.SplitSeq(s, ";") {
		// Split each part by the first "=" to separate the key and value.
		kv := strings.SplitN(part, "=", 2)
		// If the split does not result in exactly two parts (key and value),
		// return an error indicating invalid format.
		if len(kv) != 2 {
			return nil, tserr.InvalidFormat(fmt.Sprintf("invalid RRULE part: %s", part))
		}
		// Use a switch statement to handle different keys in the RRULE.
		switch kv[0] {
		// If the key is "FREQ", assign the value to the Freq field of the RRule struct.
		case "FREQ":
			rrule.Freq = kv[1]
		// If the key is "BYMONTH", convert the value to an integer and
		// assign it to the ByMonth field.
		case "BYMONTH":
			// Convert the value to an integer representing the month (1-12).
			month, err := strconv.Atoi(kv[1])
			// If there is an error converting the value to an integer,
			// return an error indicating invalid format.
			if err != nil {
				return nil, tserr.InvalidFormat(fmt.Sprintf("invalid RRULE part: %s", part))
			}
			// Assign the parsed month value to the ByMonth field of the RRule struct.
			rrule.ByMonth = month
		// If the key is "BYDAY", assign the value to the ByDay field of the RRule struct.
		case "BYDAY":
			rrule.ByDay = kv[1]
		}
	}
	// Return the parsed RRule struct and a nil error, indicating successful parsing.
	return &rrule, nil
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
				// If the UNTIL value is not a valid date, try parsing it as a date-only value.
				if _, err := parseICSDate(kv[1]); err != nil {
					return tserr.InvalidFormat(fmt.Sprintf("invalid RRULE UNTIL value: %s", kv[1]))
				}
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
	// Return nil if all checks pass to indicate a valid RRULE.
	return nil
}
