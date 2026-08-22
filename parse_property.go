// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

// Import the tserr package to handle error reporting and formatting in the functions.
import (
	"fmt"     // For formatting error messages.
	"strings" // For string manipulation.
	"time"    // For time operations.

	"github.com/thorsphere/tserr" // For error handling.
)

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

// parseICSDate parses an ICS date-only value ("20250301") into a
// time.Time at midnight UTC. Per RFC 5545 §3.3.4, DATE values are
// floating; this package normalizes them to midnight UTC.
func parseICSDate(s string) (time.Time, error) {
	// time.Parse returns UTC when the layout has no zone indicator,
	// and the date-only layout has no time components, so the result
	// is midnight UTC by construction.
	t, err := time.Parse("20060102", s)
	// If there is an error during parsing, return a zero time.Time value and
	// an error indicating invalid format.
	if err != nil {
		return time.Time{}, tserr.InvalidFormat(fmt.Sprintf("invalid date format: %s", s))
	}
	// Return the parsed time.Time value and a nil error, indicating successful parsing.
	return t, nil
}

// splitKeyValue splits a line into a key-value pair based on the first colon.
func splitKeyValue(line string) (*keyValue, error) {
	// Split the line into key-value pairs based on the first colon.
	parts := strings.SplitN(line, ":", 2)
	// If the line does not contain a colon, return an error indicating invalid format.
	if len(parts) != 2 {
		return nil, tserr.InvalidFormat(line)
	}
	// Extract the key and value from the split parts.
	key := parts[0]
	// The value is the part after the colon.
	value := parts[1]
	// Return a new keyValue struct containing the extracted key and value.
	return &keyValue{Key: key, Value: value}, nil
}
