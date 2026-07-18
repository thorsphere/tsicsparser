// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import necessary packages for testing and ICS parsing.
import (
	"strings"
	"testing" // Import the testing package for writing test cases.

	"github.com/thorsphere/tserr"       // Import the tserr package for error handling.
	"github.com/thorsphere/tsfio"       // Import the tsfio package for file handling utilities.
	"github.com/thorsphere/tsicsparser" // Import the tsicsparser package to test the parseTimezone function.
)

// TestParseTimezone tests the parseTimezone function by reading a sample ICS file
// containing timezone information and comparing the output against a golden file.
// It ensures that the parser correctly extracts timezone data and produces the expected output.
func TestParseTimezone(t *testing.T) {
	// Define the path to the ICS file that will be used for testing.
	fn := "testdata/timezone.ics"
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
	// Use a variable to hold the parsed timezone information.
	var tz tsicsparser.Timezone
	// Use a variable to hold any error that occurs during parsing.
	var err error
	// Scan through the ICS file to find the "BEGIN:VTIMEZONE" line
	// and parse the timezone information.
	for s.Scan() {
		// Read the current line from the scanner.
		line := s.Text()
		// If we find the "BEGIN:VTIMEZONE" line, we call the ParseTimezone function
		if line == "BEGIN:VTIMEZONE" {
			// If we find the "BEGIN:VTIMEZONE" line, we call the ParseTimezone function
			tz, err = tsicsparser.ParseTimezone(s)
			// If there is an error parsing the timezone, we report it and stop the test.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseTimezone", Fn: fn, Err: err}))
			}
			// If we successfully parsed the timezone, we can break out of the loop.
			break
		}
	}
	// Check for any errors that occurred during scanning and report them.
	if err := s.Err(); err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Scan", Fn: fn, Err: err}))
	}
	// Evaluate the parsed timezone against the golden file to ensure correctness.
	if err := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "timezone", Data: tz.String()}); err != nil {
		t.Fatal(err)
	}
}

// TestTimezoneWithoutRules tests the Timezone struct with a TZID of "Test" and no rules.
// It ensures that the Timezone struct correctly handles this scenario and produces the expected output.
func TestTimezoneWithoutRules(t *testing.T) {
	// Create a new Timezone struct with a TZID of "Test" and no rules.
	tz := tsicsparser.Timezone{
		TZID: "Test",
	}
	// Evaluate the parsed timezone against the golden file to ensure correctness.
	if err := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "timezone-wr", Data: tz.String()}); err != nil {
		t.Fatal(err)
	}
}

// TestParseTimezoneErrors tests the parseTimezone function with various malformed ICS inputs
// to ensure that it correctly identifies and reports parsing errors. Each test case is designed
// to trigger a specific error condition in the parser.
func TestParseTimezoneErrors(t *testing.T) {
	// Define a slice of test cases, each containing a name and an input string representing
	// a malformed timezone definition in ICS format. These test cases are designed to trigger
	// specific parsing errors in the ParseTimezone function.
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "END_DAYLIGHT_without_BEGIN",
			input: "BEGIN:VTIMEZONE\n\nTZID:Test\nEND:DAYLIGHT\nEND:VTIMEZONE",
		},
		{
			name:  "END_STANDARD_without_BEGIN",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nEND:STANDARD\nEND:VTIMEZONE",
		},
		{
			name:  "DTSTART_outside_rule",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nDTSTART:20070311T020000\nEND:VTIMEZONE",
		},
		{
			name:  "RRULE_outside_rule",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nRRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU\nEND:VTIMEZONE",
		},
		{
			name:  "TZOFFSETFROM_outside_rule",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nTZOFFSETFROM:-0400\nEND:VTIMEZONE",
		},
		{
			name:  "TZOFFSETTO_outside_rule",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nTZOFFSETTO:-0500\nEND:VTIMEZONE",
		},
		{
			name:  "TZNAME_outside_rule",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nTZNAME:Test\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_offset_from",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:STANDARD\nTZOFFSETFROM:bad\nEND:STANDARD\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_offset_to",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:STANDARD\nTZOFFSETTO:bad\nEND:STANDARD\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_offset_to_2",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:STANDARD\nTZOFFSETTO:+bado\nEND:STANDARD\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_offset_to_3",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:STANDARD\nTZOFFSETTO:badof\nEND:STANDARD\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_offset_to_4",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:STANDARD\nTZOFFSETTO:-04ba\nEND:STANDARD\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_dtstart",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:STANDARD\nDTSTART:bad\nEND:STANDARD\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_begin_rule",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:DAYLIGT\nTZOFFSETTO:-0400\nEND:DAYLIGHT\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_end_rule",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:DAYLIGHT\nTZOFFSETTO:-0400\nEND:DAYLIGT\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_rrule_1",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:DAYLIGHT\nRRULE:\nEND:DAYLIGHT\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_rrule_2",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:DAYLIGHT\nRRULE:invalid\nEND:DAYLIGHT\nEND:VTIMEZONE",
		},
		{
			name:  "invalid_rrule_3",
			input: "BEGIN:VTIMEZONE\nTZID:Test\nBEGIN:DAYLIGHT\nRRULE:FREQ=YEARLY;BYMONTH=bad;BYDAY=1SU\nEND:DAYLIGHT\nEND:VTIMEZONE",
		},
		{
			name:  "missing_END_VTIMEZONE",
			input: "BEGIN:VTIMEZONE\nTZID:Test",
		},
		{
			name:  "invalid_TZID",
			input: "BEGIN:VTIMEZONE\nTZID;Test",
		},
	}
	// Iterate over each test case defined in the tests slice.
	for _, tt := range tests {
		// Run each test case as a subtest to isolate failures and provide better reporting.
		t.Run(tt.name, func(t *testing.T) {
			// Create a new ICSScanner with the test input string.
			s := tsicsparser.NewICSScanner(strings.NewReader(tt.input), "test")
			// Scan the input to prepare for parsing.
			s.Scan()
			// Call the ParseTimezone function with the test input and check for errors.
			_, err := tsicsparser.ParseTimezone(s)
			// Check if the error is nil, which indicates that the parser did not catch the expected error.
			if err == nil {
				t.Fatal(tserr.NilFailed(tt.name))
			}
		})
	}
}
