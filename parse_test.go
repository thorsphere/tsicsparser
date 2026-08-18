// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import necessary packages for testing the tsicsparser package.
import (
	"strings" // Import the strings package to work with string manipulation functions.
	"testing" // Import the testing package to write unit tests for the tsicsparser package.
	"time"    // Import the time package to work with date and time values in the tests.

	"github.com/thorsphere/tserr"       // Import the tserr package to handle error reporting and formatting in the tests.
	"github.com/thorsphere/tsfio"       // Import the tsfio package to handle file input/output operations in the tests.
	"github.com/thorsphere/tsicsparser" // Import the tsicsparser package to test the parseTimezone function.
)

// TestParseProdId tests the parseProdId function by providing a sample product identifier string
// and verifying that the parsed output matches the expected values. It ensures that the parser
// correctly extracts the registered flag, organization, product, and language components from
// the product identifier string.
func TestParseProdId(t *testing.T) {
	// Define the expected components of the product identifier.
	r := "+"
	o := "Thorsphere"
	p := "tsicsparser"
	l := "en"
	// Define a sample product identifier string to test the parseProdId function.
	pid := r + "//" + o + "//" + p + "//" + l
	// Call the parseProdId function to parse the product identifier string.
	prodId, err := tsicsparser.ParseProdId(pid)
	// If there is an error during parsing, report it and stop the test.
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: pid, Err: err}))
	}
	// Check if the parsed product identifier is registered (the first component should be "+").
	if !prodId.Registered {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: pid, Err: tserr.InvalidFormat("Expected registered ProdID")}))
	}
	// Check if the organization field matches the expected value.
	if prodId.Organisation != o {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: prodId.Organisation, Err: tserr.InvalidFormat("Unexpected Organisation")}))
	}
	// Check if the product field matches the expected value.
	if prodId.Product != p {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: prodId.Product, Err: tserr.InvalidFormat("Unexpected Product")}))
	}
	// Check if the language field matches the expected value.
	if prodId.Language != l {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: prodId.Language, Err: tserr.InvalidFormat("Unexpected Language")}))
	}
}

// TestParseProdIdErr tests the parseProdId function with invalid product identifier strings
// to ensure that it correctly returns errors for malformed input. It verifies that the parser
// handles cases where the registered flag or product identifier format is incorrect.
func TestParseProdIdErr(t *testing.T) {
	// Define a set of test cases with invalid calendar input strings to test error handling.
	tests := []struct {
		name  string
		input string
	}{
		{name: "Invalid Registered Flag", input: "bad//Thorsphere//tsicsparser//en"},
		{name: "Invalid ProdID", input: "+//Thorsphere//tsicsparser"},
	}
	// Iterate over the test cases and run each one as a subtest.
	for _, tt := range tests {
		// Run each test case as a subtest to isolate failures and provide better reporting.
		t.Run(tt.name, func(t *testing.T) {
			// Call the parseProdId function to parse the product identifier string.
			_, err := tsicsparser.ParseProdId(tt.input)
			// If there is an error during parsing, report it and stop the test.
			if err == nil {
				t.Fatal(tserr.NilFailed(tt.name))
			}
		},
		)
	}
}

// TestParseCalendarErr tests the ParseCalendarErr function.
// It defines a set of test cases with invalid calendar input strings
// to ensure that it correctly returns errors for malformed input.
func TestParseCalendarErr(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Tests the parsing of a calendar that is missing the required PRODID field.
		// It verifies that the parser correctly identifies the missing field.
		{name: "missing PRODID", input: "BEGIN:VCALENDAR\nVERSION:2.0\nEND:VCALENDAR"},
		// Tests the parsing of a calendar that is missing the required VERSION field.
		// It verifies that the parser correctly identifies the missing field.
		{name: "missing VERSION", input: "BEGIN:VCALENDAR\nPRODID:-//Thorsphere//tsicsparser//en\nEND:VCALENDAR"},
		// Tests that a second BEGIN:VCALENDAR appearing before the first END:VCALENDAR is rejected,
		// rather than silently parsed as part of the outer calendar.
		{name: "nested BEGIN:VCALENDAR should be rejected", input: "BEGIN:VCALENDAR\n" +
			"VERSION:2.0\n" +
			"PRODID:+//Thorsphere//tsicsparser//en\n" +
			"BEGIN:VCALENDAR\n" +
			"VERSION:2.0\n" +
			"PRODID:+//Thorsphere//inner//en\n" +
			"END:VCALENDAR\n" +
			"END:VCALENDAR"},
	}
	// Iterate over the test cases and run each one as a subtest.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create a new ICSScanner to read the input string.
			s := tsicsparser.NewICSScanner(strings.NewReader(tt.input), "test")
			// Call the ParseCalendar function to parse the calendar input string.
			_, err := tsicsparser.ParseCalendar(s)
			// Check if the error is nil, indicating that the parser did not detect the missing PRODID field.
			if err == nil {
				// If the error is nil, it means that the parser did not detect the error.
				// Therefore, we call t.Fatal to indicate that the test has failed and provide an appropriate error message.
				t.Fatal(tserr.NilFailed(tt.name))
			}
		},
		)
	}
}

// TestParseCalendarEventBeforeTimezone tests the parsing of a calendar event that
// appears before its associated timezone definition.
// It ensures that the parser correctly resolves the event's start time based
// on the timezone information provided later in the input.
func TestParseCalendarEventBeforeTimezone(t *testing.T) {
	// VEVENT appears BEFORE its VTIMEZONE. Pre-fix this failed with
	// "no rules in timezone"; post-fix the event resolves correctly.
	input := "BEGIN:VCALENDAR\n" +
		"VERSION:2.0\n" +
		"PRODID:+//Thorsphere//tsicsparser//en\n" +
		"BEGIN:VEVENT\n" +
		"UID:tz-after-event@test\n" +
		"DTSTART;TZID=US-Eastern:20250102T100000\n" +
		"DTEND;TZID=US-Eastern:20250102T110000\n" +
		"END:VEVENT\n" +
		"BEGIN:VTIMEZONE\n" +
		"TZID:US-Eastern\n" +
		"BEGIN:DAYLIGHT\n" +
		"TZOFFSETFROM:-0500\n" +
		"TZOFFSETTO:-0400\n" +
		"DTSTART:20250309T020000\n" +
		"RRULE:FREQ=YEARLY;BYMONTH=3;BYDAY=2SU\n" +
		"END:DAYLIGHT\n" +
		"BEGIN:STANDARD\n" +
		"TZOFFSETFROM:-0400\n" +
		"TZOFFSETTO:-0500\n" +
		"DTSTART:20251102T020000\n" +
		"RRULE:FREQ=YEARLY;BYMONTH=11;BYDAY=1SU\n" +
		"END:STANDARD\n" +
		"END:VTIMEZONE\n" +
		"END:VCALENDAR"
	// Create a new ICSScanner to read the input string.
	s := tsicsparser.NewICSScanner(strings.NewReader(input), "test")
	// Call the ParseCalendar function to parse the calendar input string.
	cal, err := tsicsparser.ParseCalendar(s)
	// If there is an error during parsing, report it and stop the test.
	if err != nil {
		// Report an error if there was an issue parsing the calendar.
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseCalendar", Err: err}))
	}
	// Check if the parsed calendar contains exactly one event.
	if len(cal.Events) != 1 {
		// Report an error if the number of events in the parsed calendar does not match the expected count.
		t.Fatal(tserr.EqualInt(&tserr.EqualIntArgs{Var: "number of events", Actual: int64(len(cal.Events)), Want: 1}))
	}
	// Define the expected UTC time for the event's start time.
	want := time.Date(2025, time.January, 2, 15, 0, 0, 0, time.UTC) // 10:00 EST -> 15:00 UTC
	// Check if the parsed event's start time matches the expected UTC time.
	if !cal.Events[0].Start.Equal(want) {
		// Report an error if the actual start time does not match the expected time.
		t.Error(tserr.EqualStr(&tserr.EqualStrArgs{Var: "event start", Actual: cal.Events[0].Start.Format(time.RFC3339), Want: want.Format(time.RFC3339)}))
	}
}

// TestParseCal tests the parsing of a calendar from an ICS file.
// It reads the ICS file, parses the calendar data, and compares the output against a golden file
// to ensure that the parsing is correct and consistent with expected results.
func TestParseCal(t *testing.T) {
	// Define the path to the ICS file that will be used for testing.
	fn := "testdata/cal.ics"
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
	// Call the ParseCalendar function to parse the calendar data from the scanner.
	cal, err := tsicsparser.ParseCalendar(s)
	// If there is an error during parsing, report it and stop the test.
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseCalendar", Fn: fn, Err: err}))
	}
	// Check for any errors that occurred during scanning and report them.
	if err := s.Err(); err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Scan", Fn: fn, Err: err}))
	}
	// Evaluate the parsed timezone against the golden file to ensure correctness.
	if err := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "cal", Data: cal.String()}); err != nil {
		t.Fatal(err)
	}
}

// TestParseCalendarName tests the parsing of a calendar with a NAME field.
// It verifies that the parser correctly extracts the calendar name from the input string.
func TestParseCalendarName(t *testing.T) {
	// Define tests for different calendar name scenarios, including cases with NAME and SUMMARY fields.
	tests := []struct {
		name  string
		input string
	}{
		{ // Test case with a NAME field, expecting the calendar name to be "Standard Name".
			name: "Standard Name",
			input: "BEGIN:VCALENDAR\n" +
				"VERSION:2.0\n" +
				"PRODID:+//Thorsphere//tsicsparser//en\n" +
				"NAME:Standard Name\n" +
				"SUMMARY:Legacy Summary\n" +
				"END:VCALENDAR",
		},
		{ // Test case with only a SUMMARY field, expecting the name to be derived from the SUMMARY.
			name: "Legacy Summary",
			input: "BEGIN:VCALENDAR\n" +
				"VERSION:2.0\n" +
				"PRODID:+//Thorsphere//tsicsparser//en\n" +
				"SUMMARY:Legacy Summary\n" +
				"END:VCALENDAR",
		},
		{ // Test case with no NAME or SUMMARY fields, expecting an empty name.
			name: "",
			input: "BEGIN:VCALENDAR\n" +
				"VERSION:2.0\n" +
				"PRODID:+//Thorsphere//tsicsparser//en\n" +
				"X-WR-CALNAME:My Calendar\n" +
				"END:VCALENDAR",
		},
	}
	// Iterate over the test cases and run each one as a subtest.
	for _, tt := range tests {
		// Run each test case as a subtest to isolate failures and provide better reporting.
		t.Run(tt.name, func(t *testing.T) {
			// Create a new ICSScanner to read the input string.
			s := tsicsparser.NewICSScanner(strings.NewReader(tt.input), "test")
			// Call the ParseCalendar function to parse the calendar input string.
			cal, err := tsicsparser.ParseCalendar(s)
			// If there is an error during parsing, report it and stop the test.
			if err != nil {
				// Report an error if there was an issue parsing the calendar.
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseCalendar", Err: err}))
			}
			// Check if the parsed calendar name matches the expected name for the test case.
			if cal.Name != tt.name {
				// Report an error if the actual calendar name does not match the expected name.
				t.Fatal(tserr.EqualStr(&tserr.EqualStrArgs{
					Var: "Name", Actual: cal.Name, Want: tt.name,
				}))
			}
		})
	}
}

// TestCollectRawBlockErrors tests the collectRawBlock function with
// various malformed inputs to ensure that it correctly identifies and
// reports structural errors. Each test case is designed to trigger a
// specific error condition in the CollectRawBlock function:
//   - a line that cannot be split into a key-value pair (no colon),
//   - an END:<other> at depth 0 with no matching BEGIN (unexpected END),
//   - EOF before the matching END:<block> (NotFound).
func TestCollectRawBlockErrors(t *testing.T) {
	// Define a slice of test cases, each containing a name and an input
	// string representing a malformed VEVENT body in ICS format. The
	// "BEGIN:VEVENT" line is consumed by the caller (mirroring
	// parseCalendar), so each body starts directly with the first
	// property line.
	tests := []struct {
		name  string
		input string
	}{
		{
			// A line with no colon cannot be split into a key-value pair
			// by splitKeyValue, so collectRawBlock returns that error.
			name:  "invalid_line_no_colon",
			input: "DTSTART:20250101T120000Z\nSUMMARY-Test\nEND:VEVENT",
		},
		{
			// An END:<other> at depth 0 (no matching BEGIN:<other>) is a
			// structural error: "unexpected END:VTODO inside VEVENT".
			name:  "unexpected_end_other",
			input: "DTSTART:20250101T120000Z\nEND:VTODO",
		},
		{
			// EOF before the matching END:VEVENT. The stream ends without
			// the close tag, so collectRawBlock returns NotFound.
			name:  "missing_END_VEVENT",
			input: "DTSTART:20250101T120000Z\nSUMMARY:Test",
		},
	}
	// Iterate over each test case defined in the tests slice.
	for _, tt := range tests {
		// Run each test case as a subtest to isolate failures and provide better reporting.
		t.Run(tt.name, func(t *testing.T) {
			// Create a new ICSScanner with the test input string.
			s := tsicsparser.NewICSScanner(strings.NewReader(tt.input), "test")
			// Call the CollectRawBlock function with block "VEVENT" and
			// check for errors.
			_, err := tsicsparser.CollectRawBlock(s, "VEVENT")
			// Check if the error is nil, which indicates that the parser
			// did not catch the expected error.
			if err == nil {
				t.Fatal(tserr.NilFailed(tt.name))
			}
		})
	}
}

// TestParseBufferedEventsErr tests that a parseEvent error during pass 2
// is surfaced by parseBufferedEvents through its error branch, rather
// than being silently swallowed. Each case provides a raw VEVENT body
// (the lines between BEGIN:VEVENT and END:VEVENT, exclusive, as
// collected by pass 1) that makes parseEvent return an error when
// parseBufferedEvents reconstructs and re-scans it.
func TestParseBufferedEventsErr(t *testing.T) {
	// Define test cases, each providing a raw VEVENT body that triggers
	// a distinct parseEvent failure during pass 2.
	tests := []struct {
		name string
		raw  []string
	}{
		{
			// No DTSTART: parseEvent reaches END:VEVENT with a zero
			// Start (and no End/DURATION) and returns
			// "missing required DTSTART in VEVENT".
			name: "missing required DTSTART",
			raw:  []string{"SUMMARY:Test"},
		},
		{
			// A line without a colon cannot be split by splitKeyValue,
			// so parseEvent returns that error before reaching END:VEVENT.
			name: "line without colon",
			raw:  []string{"DTSTART:20250101T120000Z", "INVALIDLINE"},
		},
	}
	// Iterate over each test case and run it as a subtest.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A minimal calendar; Timezones is empty because the
			// malformed events do not reference any TZID.
			cal := &tsicsparser.Calendar{}
			// Call ParseBufferedEvents with a single malformed raw block.
			// parseBufferedEvents joins the raw lines and appends
			// "\nEND:VEVENT" before handing the body to parseEvent.
			err := tsicsparser.ParseBufferedEvents(cal, [][]string{tt.raw})
			// If there is no error, the parseEvent error was not
			// propagated through the "return err" branch — the test
			// has failed.
			if err == nil {
				t.Fatal(tserr.NilFailed(tt.name))
			}
		})
	}
}
