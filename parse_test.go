// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

import (
	"strings"
	"testing"
	"time"

	"github.com/thorsphere/tserr"
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
	// Define a set of test cases with invalid product identifier strings to test error handling.
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
				t.Fatal(tserr.NilFailed("parseProdID"))
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
