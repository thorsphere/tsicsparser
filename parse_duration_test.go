// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import the necessary packages for testing, time handling, and custom error handling.
import (
	"fmt"     // For formatting error messages.
	"testing" // For writing test cases.
	"time"    // For handling time durations.

	"github.com/thorsphere/tserr"       // For error handling.
	"github.com/thorsphere/tsicsparser" // For testing the tsicsparser package.
)

// TestParseDuration tests the parseDuration function.
// It checks that valid ICS duration strings (RFC 5545 §3.3.6) are parsed
// into the expected time.Duration values.
func TestParseDuration(t *testing.T) {
	// Define a test case for each duration form.
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"PT0M", 0},
		{"PT1H30M", 1*time.Hour + 30*time.Minute},
		{"P1DT2H", 24*time.Hour + 2*time.Hour},
		{"P2W", 2 * 7 * 24 * time.Hour},
		{"P1D", 24 * time.Hour},
		{"PT45S", 45 * time.Second},
		{"PT1M", time.Minute},
		{"PT1H", time.Hour},
		{"-PT1H", -1 * time.Hour},
		{"+PT15M", 15 * time.Minute},
		{"P1DT1H1M1S", 24*time.Hour + time.Hour + time.Minute + time.Second},
	}
	// Loop through the tests and check if the duration matches the expected value.
	for _, tt := range tests {
		// Call the parseDuration function with the test input.
		dur, err := tsicsparser.ParseDuration(tt.input)
		// If there is an error parsing the duration, report it and continue to the next test case.
		if err != nil {
			t.Error(tserr.Op(&tserr.OpArgs{Op: "ParseDuration", Fn: tt.input, Err: err}))
		}
		// Check if the parsed duration matches the expected value.
		if dur != tt.want {
			// If the duration does not match, report an error with details.
			t.Error(tserr.EqualInt(&tserr.EqualIntArgs{Var: fmt.Sprintf("duration %s", tt.input), Want: int64(tt.want), Actual: int64(dur)}))
		}
	}
}

// TestParseDurationErr tests the parseDuration function.
// It checks that invalid duration strings return an error.
func TestParseDurationErr(t *testing.T) {
	// Define invalid duration strings that should be rejected.
	invalid := []string{
		"",                   // empty
		"P",                  // no components
		"PT",                 // no components
		"-P",                 // no components
		"PD",                 // no digits
		"P1",                 // no unit
		"1H",                 // missing 'P'
		"PT1X",               // invalid unit
		"P1H",                // 'H' before 'T'
		"PT1D",               // 'D' in time section
		"P99999999W",         // overflow
		"P106751DT23H47M17S", // overflow
	}
	// Loop through the invalid inputs and check that they return an error.
	for _, in := range invalid {
		// Call the parseDuration function with the invalid input.
		if _, err := tsicsparser.ParseDuration(in); err == nil {
			// If no error is returned for an invalid input, report a failure.
			t.Error(tserr.NilFailed("ParseDuration"))
		}
	}
}
