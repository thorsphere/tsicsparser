// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import the necessary packages for testing, string manipulation, and custom error handling.
import (
	"fmt"     // For formatting error messages.
	"strings" // For string manipulation.
	"testing" // For writing test cases.

	"github.com/thorsphere/lpstats"     // For comparing string maps in tests.
	"github.com/thorsphere/tserr"       // For error handling.
	"github.com/thorsphere/tsfio"       // For file handling utilities.
	"github.com/thorsphere/tsicsparser" // For testing the tsicsparser package.
)

// TestParseEvent tests the ParseEvent function by reading a sample ICS file
// containing events and comparing the output against a golden file.
// It ensures that the parser correctly extracts event data and produces the expected output.
func TestParseEvent(t *testing.T) {
	// Define the path to the ICS file that will be used for testing.
	fn := "testdata/events.ics"
	// Open the ICS file for reading using the tsfio package, which provides file handling utilities.
	f, e := tsfio.OpenFile(tsfio.Filename(fn))
	// If there is an error opening the file, we report it and stop the test.
	if e != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "OpenFile", Fn: fn, Err: e}))
	}
	// Ensure the file is closed after the test completes to avoid resource leaks.
	defer f.Close()
	// Create a new ICSScanner to read from the opened file.
	s := tsicsparser.NewICSScanner(f, fn)
	// Use variables to hold the parsed event information.
	var tz tsicsparser.Timezone
	var ev tsicsparser.Event
	var sb strings.Builder
	// Use a variable to hold any error that occurs during parsing.
	var err error
	// Scan through the ICS file to find the "BEGIN:VTIMEZONE" line
	// and parse the timezone information.
	for s.Scan() {
		// Read the current line from the scanner.
		line := s.Text()
		// Switch on the current line to determine what to do.
		switch line {
		// If we find the "BEGIN:VTIMEZONE" line, we call the ParseTimezone function
		case "BEGIN:VTIMEZONE":
			// If we find the "BEGIN:VTIMEZONE" line, we call the ParseTimezone function
			tz, err = tsicsparser.ParseTimezone(s)
			// If there is an error parsing the timezone, we report it and stop the test.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseTimezone", Fn: fn, Err: err}))
			}
			// Write the parsed timezone to the output buffer.
			sb.WriteString(tz.String())
		case "BEGIN:VEVENT":
			// If we find the "BEGIN:VEVENT" line, we call the ParseEvent function
			ev, err = tsicsparser.ParseEvent(s, tz)
			// If there is an error parsing the event, we report it and stop the test.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseEvent", Fn: fn, Err: err}))
			}
			// Write the parsed event to the output buffer.
			sb.WriteString(ev.String())
		default:
			// If we find any other line, we skip it.
		}
	}
	// Check for any errors that occurred during scanning and report them.
	if err := s.Err(); err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Scan", Fn: fn, Err: err}))
	}
	tsfio.CreateGoldenFile(&tsfio.Testcase{Name: "events", Data: sb.String()})
	// Evaluate the parsed timezone against the golden file to ensure correctness.
	if err := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "events", Data: sb.String()}); err != nil {
		t.Fatal(err)
	}
}

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
