// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import the necessary packages for testing, string manipulation, and error handling.
import (
	"fmt"     // For formatting error messages.
	"testing" // For writing test cases.
	"time"    // For time operations.

	"github.com/thorsphere/lpstats"     // For comparing string maps in tests.
	"github.com/thorsphere/tserr"       // For error handling.
	"github.com/thorsphere/tsicsparser" // For testing the tsicsparser package.
)

// TestSplitKeyParams tests the SplitKeyParams function.
// It checks that the function correctly splits a key string into its base
// name and a map of parameters. It fails if the function returns an error
// or if the parsed base/params don't match the expected values.
func TestSplitKeyParams(t *testing.T) {
	// Define a set of test cases covering the common shapes of ICS keys.
	tests := []struct {
		name       string
		input      string
		wantBase   string
		wantParams map[string]string
	}{
		{
			name:       "no parameters",
			input:      "DTSTART",
			wantBase:   "DTSTART",
			wantParams: map[string]string{},
		},
		{
			name:       "single parameter",
			input:      "DTSTART;TZID=America/New_York",
			wantBase:   "DTSTART",
			wantParams: map[string]string{"TZID": "America/New_York"},
		},
		{
			name:       "single parameter with double quotes",
			input:      "DTSTART;TZID=\"America/New_York\"",
			wantBase:   "DTSTART",
			wantParams: map[string]string{"TZID": "America/New_York"},
		},
		{
			name:       "multiple parameters",
			input:      "DTSTART;TZID=US-Eastern;VALUE=DATE-TIME",
			wantBase:   "DTSTART",
			wantParams: map[string]string{"TZID": "US-Eastern", "VALUE": "DATE-TIME"},
		},
		{
			name:       "empty parameter value",
			input:      "DTSTART;TZID=",
			wantBase:   "DTSTART",
			wantParams: map[string]string{"TZID": ""},
		},
	}
	// Run each test case as a subtest for clearer failure attribution.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Split the key into base name and parameters.
			base, params, err := tsicsparser.SplitKeyParams(tt.input)
			// Fail immediately if an unexpected error is returned.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "splitKeyParams", Err: err}))
			}
			// Check that the base name matches the expected value.
			if base != tt.wantBase {
				t.Error(tserr.EqualStr(&tserr.EqualStrArgs{
					Var:    "base",
					Want:   tt.wantBase,
					Actual: base,
				}))
			}
			// Check that the parameter map matches the expected map.
			if !lpstats.EqualStrMaps(params, tt.wantParams) {
				t.Error(tserr.EqualStr(&tserr.EqualStrArgs{
					Var:    "params",
					Want:   fmt.Sprintf("%v", tt.wantParams),
					Actual: fmt.Sprintf("%v", params),
				}))
			}
		})
	}
}

// TestSplitKeyParamsErr tests the SplitKeyParams function.
// It checks that the function returns an error for parameter strings that
// don't contain '='. It fails if the function returns nil.
func TestSplitKeyParamsErr(t *testing.T) {
	// Define a set of test cases for malformed parameter formats.
	tests := []struct {
		name  string
		input string
	}{
		{"trailing parameter without equals", "DTSTART;INVALID"},
		{"second parameter without equals", "DTSTART;TZID=US;ANOTHER"},
		{"only semicolon", "DTSTART;"},
	}
	// Loop through the tests and verify that each one returns an error.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Split the key into base name and parameters.
			_, _, err := tsicsparser.SplitKeyParams(tt.input)
			// Verify that an error is returned for the malformed input.
			if err == nil {
				t.Error(tserr.NilFailed("splitKeyParams"))
			}
		})
	}
}

// TestParseICSDate tests the parseICSDate function.
// It checks that an ICS date-only value ("20250301") parses into a
// time.Time at exactly midnight UTC — both the instant and the location.
// time.Parse returns UTC when the layout has no zone indicator, and the
// date-only layout has no time components, so midnight UTC holds by
// construction; this test pins that invariant against future refactors
// (e.g. a switch to ParseInLocation would silently break it).
func TestParseICSDate(t *testing.T) {
	// Define a set of test cases covering valid date-only values,
	// including calendar edge cases (leap day, year boundary).
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{
			name:  "regular date",
			input: "20250301",
			want:  time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "leap day",
			input: "20240229",
			want:  time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "year boundary",
			input: "20250101",
			want:  time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	// Run each test case as a subtest for clearer failure attribution.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the date-only value into a time.Time.
			got, err := tsicsparser.ParseICSDate(tt.input)
			// Fail immediately if an unexpected error is returned.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseICSDate", Fn: tt.input, Err: err}))
			}
			// Check the instant: midnight (00:00:00) on the parsed date.
			if !got.Equal(tt.want) {
				t.Error(tserr.EqualStr(&tserr.EqualStrArgs{
					Var:    "date",
					Want:   tt.want.Format(time.RFC3339),
					Actual: got.Format(time.RFC3339),
				}))
			}
			// Check the location: must be UTC, not just an equal instant.
			// Equal() alone cannot catch a location change, because it
			// compares instants regardless of the attached location.
			if got.Location() != time.UTC {
				t.Error(tserr.EqualStr(&tserr.EqualStrArgs{
					Var:    "location",
					Want:   time.UTC.String(),
					Actual: got.Location().String(),
				}))
			}
			// Check the clock fields explicitly: hour, minute, second and
			// nanosecond must all be zero — the "midnight" half of the
			// invariant, spelled out so a failure reads unambiguously.
			if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
				t.Error(tserr.EqualStr(&tserr.EqualStrArgs{
					Var:    "clock",
					Want:   "00:00:00.000000000",
					Actual: got.Format("15:04:05.000000000"),
				}))
			}
		})
	}
}

// TestParseICSDateErr tests the parseICSDate function.
// It checks that malformed date-only values return an error rather than
// being silently coerced into a plausible date.
func TestParseICSDateErr(t *testing.T) {
	// Define a set of malformed inputs. Note the datetime-shaped entry:
	// a full DATE-TIME value must NOT be accepted by the date-only parser,
	// otherwise DTSTART;VALUE=DATE:20250301T120000 would parse instead of
	// erroring.
	tests := []struct {
		name  string
		input string
	}{
		{name: "not a date", input: "notadate"},
		{name: "empty string", input: ""},
		{name: "too short", input: "2025030"},
		{name: "too long", input: "202503011"},
		{name: "invalid month", input: "20251301"},
		{name: "invalid day", input: "20250132"},
		{name: "non-leap Feb 29", input: "20250229"},
		{name: "datetime value rejected", input: "20250301T120000"},
		{name: "zulu datetime value rejected", input: "20250301T120000Z"},
	}
	// Loop through the tests and verify that each one returns an error.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the date-only value; an error is expected.
			_, err := tsicsparser.ParseICSDate(tt.input)
			// Verify that an error is returned for the invalid input.
			if err == nil {
				t.Error(tserr.NilFailed(tt.name))
			}
		})
	}
}
