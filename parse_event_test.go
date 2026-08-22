// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import the necessary packages for testing, error handling, and string manipulation.
import (
	"errors"  // For error handling.
	"strings" // For string manipulation.
	"testing" // For writing test cases.
	"time"    // For time operations.

	"github.com/thorsphere/tserr"       // For error handling.
	"github.com/thorsphere/tsfio"       // For file handling utilities.
	"github.com/thorsphere/tsicsparser" // For testing the tsicsparser package.
)

// TestParseEvent tests the ParseEvent function by reading a sample ICS file
// containing events and comparing the output against a golden file.
// It ensures that the parser correctly extracts event data and produces the expected output.
func TestParseEvent(t *testing.T) {
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
	// Use variables to hold the parsed event information.
	var tzs tsicsparser.Timezones
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
			tz, err := tsicsparser.ParseTimezone(s)
			// If there is an error parsing the timezone, we report it and stop the test.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseTimezone", Fn: fn, Err: err}))
			}
			// Collect every VTIMEZONE so events can resolve their TZID.
			tzs = append(tzs, tz)
			// Write the parsed timezone to the output buffer.
			sb.WriteString(tz.String())
		case "BEGIN:VEVENT":
			// If we find the "BEGIN:VEVENT" line, we call the ParseEvent function
			ev, err = tsicsparser.ParseEvent(s, tzs)
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
	// Evaluate the parsed timezone against the golden file to ensure correctness.
	if err := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "event", Data: sb.String()}); err != nil {
		t.Fatal(err)
	}
}

// TestParseDTValue tests the ParseDTValue function.
// It checks that the function correctly parses DTSTART/DTEND property values
// into UTC time.Time values, handling both UTC-suffixed (Zulu) datetimes and
// TZID-qualified local times that are converted to UTC via the calendar's
// timezone rules. It fails if the function returns an error or if the parsed
// time does not match the expected UTC value.
func TestParseDTValue(t *testing.T) {
	// Define a reusable US-Eastern timezone with DAYLIGHT and STANDARD rules
	// matching the testdata/timezone.ics fixture, so that local-to-UTC
	// conversions exercise the same transition logic as the golden-file test.
	useEastern := tsicsparser.Timezone{
		TZID: "US-Eastern",
		Rules: []tsicsparser.TimezoneRule{
			{
				Type:         tsicsparser.Daylight,
				TZOffsetFrom: -5 * 3600,
				TZOffsetTo:   -4 * 3600,
				DTStart:      time.Date(2025, time.March, 9, 2, 0, 0, 0, time.UTC),
				RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
			},
			{
				Type:         tsicsparser.Standard,
				TZOffsetFrom: -4 * 3600,
				TZOffsetTo:   -5 * 3600,
				DTStart:      time.Date(2025, time.November, 2, 2, 0, 0, 0, time.UTC),
				RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
			},
		},
	}
	// Define a set of test cases for valid date-time values.
	tests := []struct {
		name   string
		value  string
		params map[string]string
		tzs    tsicsparser.Timezones
		prop   string
		want   time.Time
	}{
		{
			// UTC-suffixed (Zulu) datetime: no conversion needed, TZID ignored.
			name:   "zulu datetime",
			value:  "20250101T120000Z",
			params: map[string]string{},
			tzs:    tsicsparser.Timezones{useEastern},
			prop:   "DTSTART",
			want:   time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			// TZID-qualified local time in winter (Standard/EST, UTC-5).
			// 10:00 EST → 15:00 UTC.
			name:   "TZID local time in standard period",
			value:  "20250102T100000",
			params: map[string]string{"TZID": "US-Eastern"},
			tzs:    tsicsparser.Timezones{useEastern},
			prop:   "DTSTART",
			want:   time.Date(2025, time.January, 2, 15, 0, 0, 0, time.UTC),
		},
		{
			// TZID-qualified local time in summer (Daylight/EDT, UTC-4).
			// 10:00 EDT → 14:00 UTC.
			name:   "TZID local time in daylight period",
			value:  "20250704T100000",
			params: map[string]string{"TZID": "US-Eastern"},
			tzs:    tsicsparser.Timezones{useEastern},
			prop:   "DTEND",
			want:   time.Date(2025, time.July, 4, 14, 0, 0, 0, time.UTC),
		},
		{
			// Quoted TZID parameter values are stripped by SplitKeyParams before
			// reaching ParseDTValue, so the bare (unquoted) TZID is used here.
			name:   "TZID local time with bare value (quotes already stripped)",
			value:  "20250102T100000",
			params: map[string]string{"TZID": "US-Eastern"},
			tzs:    tsicsparser.Timezones{useEastern},
			prop:   "DTSTART",
			want:   time.Date(2025, time.January, 2, 15, 0, 0, 0, time.UTC),
		},
	}
	// Run each test case as a subtest for clearer failure attribution.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the date-time value into a UTC time.Time.
			got, err := tsicsparser.ParseDTValue(tt.value, tt.params, tt.tzs, tt.prop)
			// Fail immediately if an unexpected error is returned.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseDTValue", Err: err}))
			}
			// Check that the parsed time matches the expected UTC value.
			if !got.Equal(tt.want) {
				t.Errorf("ParseDTValue(%q, %v) = %s, want %s",
					tt.value,
					tt.params,
					got.Format(time.RFC3339),
					tt.want.Format(time.RFC3339))
			}
		})
	}
}

// TestParseDTValueErr tests the ParseDTValue function.
// It checks that the function returns an error for malformed datetime values,
// floating times (no TZID and no Zulu suffix), and TZID mismatches. It fails
// if the function returns nil for any case.
func TestParseDTValueErr(t *testing.T) {
	// Reuse the same US-Eastern timezone as the success cases.
	useEastern := tsicsparser.Timezone{
		TZID: "US-Eastern",
		Rules: []tsicsparser.TimezoneRule{
			{
				Type:         tsicsparser.Daylight,
				TZOffsetFrom: -5 * 3600,
				TZOffsetTo:   -4 * 3600,
				DTStart:      time.Date(2025, time.March, 9, 2, 0, 0, 0, time.UTC),
				RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
			},
			{
				Type:         tsicsparser.Standard,
				TZOffsetFrom: -4 * 3600,
				TZOffsetTo:   -5 * 3600,
				DTStart:      time.Date(2025, time.November, 2, 2, 0, 0, 0, time.UTC),
				RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
			},
		},
	}
	// Define a set of test cases that should all return an error.
	tests := []struct {
		name   string
		value  string
		params map[string]string
		tzs    tsicsparser.Timezones
		prop   string
	}{
		{
			// Malformed datetime string that cannot be parsed by parseICSDateTime.
			name:   "invalid datetime format",
			value:  "not-a-date",
			params: map[string]string{},
			tzs:    tsicsparser.Timezones{useEastern},
			prop:   "DTSTART",
		},
		{
			// Floating time: no Zulu suffix and no TZID parameter.
			name:   "floating time without TZID",
			value:  "20250101T120000",
			params: map[string]string{},
			tzs:    tsicsparser.Timezones{useEastern},
			prop:   "DTSTART",
		},
		{
			// TZID present but does not match the timezone's TZID.
			name:   "TZID mismatch",
			value:  "20250101T120000",
			params: map[string]string{"TZID": "Europe/London"},
			tzs:    tsicsparser.Timezones{useEastern},
			prop:   "DTSTART",
		},
	}
	// Loop through the tests and verify that each one returns an error.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the date-time value; an error is expected.
			_, err := tsicsparser.ParseDTValue(tt.value, tt.params, tt.tzs, tt.prop)
			// Verify that an error is returned for the invalid input.
			if err == nil {
				t.Error(tserr.NilFailed("ParseDTValue"))
			}
		})
	}
}

// TestParseEventErr tests the parseEvent function with various malformed ICS
// inputs to ensure that it correctly identifies and reports parsing errors. Each
// test case is designed to trigger a specific error condition in parseEvent,
// mirroring the approach used by TestParseTimezoneErrors for the timezone parser.
func TestParseEventErr(t *testing.T) {
	// Define a set of test cases, each containing a name, an input string
	// representing a malformed VEVENT in ICS format, and an optional timezone.
	// The input always begins with BEGIN:VEVENT, which the caller consumes before
	// invoking parseEvent; the test harness performs that one Scan() itself.
	tests := []struct {
		name  string
		input string
		tzs   tsicsparser.Timezones
	}{
		// --- splitKeyValue / splitKeyParams errors ---
		{
			name:  "line without colon",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nINVALIDLINE\nEND:VEVENT",
		},
		{
			name:  "invalid parameter format",
			input: "BEGIN:VEVENT\nDTSTART;INVALID:20250101T120000Z\nEND:VEVENT",
		},

		// --- DTSTART duplicate ---
		{
			name:  "DTSTART already set",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nDTSTART:20250102T120000Z\nEND:VEVENT",
		},
		// --- DTEND duplicate ---
		{
			name:  "DTEND already set; duplicate DTEND",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nDTEND:20250101T130000Z\nDTEND:20250101T140000Z\nEND:VEVENT",
		},

		// --- DTEND / DURATION mutual exclusivity ---
		{
			// DURATION (after DTSTART) sets End, then DTEND hits the
			// mutual-exclusivity check on event.End.
			name:  "DTEND and DURATION mutually exclusive (DURATION then DTEND)",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nDURATION:PT1H\nDTEND:20250101T130000Z\nEND:VEVENT",
		},
		{
			// DTEND sets End, then DURATION hits the mutual-exclusivity check
			// on event.End.
			name:  "DTEND and DURATION mutually exclusive (DTEND then DURATION)",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nDTEND:20250101T130000Z\nDURATION:PT1H\nEND:VEVENT",
		},
		{
			// A pending DURATION (seen before DTSTART) counts as "DURATION set"
			// for the DTEND mutual-exclusivity check.
			name:  "DTEND and DURATION mutually exclusive (pending DURATION then DTEND)",
			input: "BEGIN:VEVENT\nDURATION:PT1H\nDTEND:20250101T130000Z\nEND:VEVENT",
		},
		// --- DURATION duplicate ---
		{
			// Two DURATIONs before DTSTART: the first is pending (End still
			// zero), so the second hits the hasDuration check, not the
			// event.End check.
			name:  "DURATION already set",
			input: "BEGIN:VEVENT\nDURATION:PT1H\nDURATION:PT2H\nEND:VEVENT",
		},
		// --- parseDTValue errors ---
		{
			name:  "invalid DTSTART datetime",
			input: "BEGIN:VEVENT\nDTSTART:bad\nEND:VEVENT",
		},
		{
			name:  "invalid DTEND datetime",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nDTEND:bad\nEND:VEVENT",
		},
		{
			// No Zulu suffix and no TZID parameter: a floating time.
			name:  "floating time without TZID",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000\nEND:VEVENT",
		},
		{
			// TZID present but does not match the calendar's TZID.
			name:  "TZID mismatch",
			input: "BEGIN:VEVENT\nDTSTART;TZID=Europe/London:20250101T120000\nEND:VEVENT",
		},
		// --- parseDuration error ---
		{
			name:  "invalid DURATION",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nDURATION:bad\nEND:VEVENT",
		},
		// --- unexpected END inside VEVENT ---
		{
			name:  "unexpected END inside VEVENT",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nEND:VALARM\nEND:VEVENT",
		},
		// --- END:VEVENT validation errors ---
		{
			// No DTSTART, no DTEND, no DURATION.
			name:  "missing required DTSTART",
			input: "BEGIN:VEVENT\nSUMMARY:Test\nEND:VEVENT",
		},
		{
			// DTEND present but DTSTART absent.
			name:  "DTEND present without DTSTART",
			input: "BEGIN:VEVENT\nDTEND:20250101T120000Z\nEND:VEVENT",
		},
		{
			// DURATION present but DTSTART absent (pending, never resolved).
			name:  "DURATION present without DTSTART",
			input: "BEGIN:VEVENT\nDURATION:PT1H\nEND:VEVENT",
		},
		// --- EOF before END:VEVENT ---
		{
			name:  "missing END:VEVENT",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z",
		},
		// --- validateRRule errors ---
		{
			// An RRULE part without "=" cannot be split into key/value,
			// so validateRRule returns "invalid RRULE part: NOEQUALS"
			// via its len(kv) != 2 branch, which parseEvent propagates
			// from its case "RRULE" arm.
			name:  "invalid RRULE part without equals",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nRRULE:FREQ=WEEKLY;NOEQUALS\nEND:VEVENT",
		},
		{
			// The RRULE's FREQ value is not one of the seven defined
			// values (SECONDLY, MINUTELY, HOURLY, DAILY, WEEKLY, MONTHLY,
			// YEARLY), so validateRRule returns "invalid RRULE FREQ
			// value: FORTNIGHTLY" via its default arm, which parseEvent
			// propagates from its case "RRULE" arm.
			name:  "invalid RRULE FREQ value",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nRRULE:FREQ=FORTNIGHTLY\nEND:VEVENT",
		},
		{
			// RFC 5545 §3.1 rule names are case-sensitive; lowercase
			// "weekly" is not a valid FREQ value.
			name:  "invalid RRULE FREQ value lowercase",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nRRULE:FREQ=weekly\nEND:VEVENT",
		},
		{
			// The RRULE's UNTIL value is not a valid ICS datetime, so
			// validateRRule returns "invalid RRULE UNTIL value: notadate"
			// via its parseICSDateTime error branch, which parseEvent
			// propagates from its case "RRULE" arm.
			name:  "invalid RRULE UNTIL value",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nRRULE:FREQ=WEEKLY;UNTIL=notadate\nEND:VEVENT",
		},
		{
			// The RRULE has no FREQ part at all. Every part parses cleanly
			// (COUNT=10 is valid), so the loop completes without error and
			// the post-loop check fires: "RRULE missing required FREQ".
			name:  "RRULE missing required FREQ",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nRRULE:COUNT=10\nEND:VEVENT",
		},
		{
			// Both UNTIL and COUNT are present. Each part is individually
			// valid (FREQ=WEEKLY is a defined value, the UNTIL datetime
			// parses), so the loop completes and the post-loop mutual-
			// exclusivity check fires: "RRULE must not contain both UNTIL
			// and COUNT" (RFC 5545 §3.3.10).
			name: "RRULE with both UNTIL and COUNT",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\n" +
				"RRULE:FREQ=WEEKLY;UNTIL=20250401T000000Z;COUNT=10\nEND:VEVENT",
		},
		{
			// An RRULE property with an empty value: splitKeyValue yields
			// Value "", which validateRRule rejects via its empty-string
			// guard before the loop is ever entered.
			name:  "empty RRULE",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nRRULE:\nEND:VEVENT",
		},
		// --- VALUE=DATE validation errors ---
		{
			// RFC 5545 §3.2.20: TZID MUST NOT be applied to DATE properties.
			// Currently the TZID parameter is silently ignored in the
			// VALUE=DATE branch; this case pins the stricter behavior.
			name:  "TZID combined with VALUE=DATE",
			input: "BEGIN:VEVENT\nDTSTART;TZID=US-Eastern;VALUE=DATE:20250301\nEND:VEVENT",
		},
		{
			// RFC 5545 §3.6.1: DTSTART and DTEND value types must match.
			// DATE start paired with a DATE-TIME end must be rejected.
			name:  "DTSTART DATE with DTEND DATE-TIME",
			input: "BEGIN:VEVENT\nDTSTART;VALUE=DATE:20250301\nDTEND:20250302T120000Z\nEND:VEVENT",
		},
		{
			// The reverse mismatch: DATE-TIME start with a DATE end.
			// Note: this case errors even without the type-match check
			// (parseDTValue rejects "20250302" as a datetime), but after
			// DTEND gains VALUE=DATE support the error must come from the
			// type-match check instead — this case guards that transition.
			name:  "DTSTART DATE-TIME with DTEND DATE",
			input: "BEGIN:VEVENT\nDTSTART:20250301T120000Z\nDTEND;VALUE=DATE:20250302\nEND:VEVENT",
		},
		{
			// RFC 5545 §3.8.2.5: with a DATE DTSTART, DURATION must be
			// dur-day or dur-week. PT1H has a time component → rejected.
			name:  "DURATION with time component after DATE DTSTART",
			input: "BEGIN:VEVENT\nDTSTART;VALUE=DATE:20250301\nDURATION:PT1H\nEND:VEVENT",
		},
		{
			// Same constraint via the pending-DURATION path: DURATION is
			// seen before DTSTART, and the check fires when DTSTART turns
			// out to be a DATE (in the DTSTART case's hasDuration branch).
			name:  "pending DURATION with time component before DATE DTSTART",
			input: "BEGIN:VEVENT\nDURATION:PT1H\nDTSTART;VALUE=DATE:20250301\nEND:VEVENT",
		},
		{
			// Two RRULE properties in one VEVENT. The first is valid and
			// stored verbatim; the second hits the duplicate check in
			// parseEvent's case "RRULE" arm ("RRULE already set") before
			// validateRRule is ever consulted for it.
			name: "RRULE already set",
			input: "BEGIN:VEVENT\nDTSTART:20250101T120000Z\n" +
				"RRULE:FREQ=WEEKLY\nRRULE:FREQ=DAILY\nEND:VEVENT",
		},
		{
			// A VALUE=DATE DTSTART whose value is not a valid ICS date.
			// The TZID check passes (no TZID present), so control reaches
			// the parseICSDate call, which rejects "20250332" (invalid
			// day-of-month) via its time.Parse error branch; parseEvent
			// propagates that error from its DTSTART case.
			name:  "invalid VALUE=DATE DTSTART",
			input: "BEGIN:VEVENT\nDTSTART;VALUE=DATE:20250332\nEND:VEVENT",
		},
		{
			// The mirror of the DTSTART case: a VALUE=DATE DTEND whose
			// value fails calendar validation. The duplicate and mutual-
			// exclusivity checks pass (first DTEND, no DURATION), so the
			// error comes from parseICSDate via the DTEND case's
			// VALUE=DATE branch.
			name:  "invalid VALUE=DATE DTEND",
			input: "BEGIN:VEVENT\nDTSTART;VALUE=DATE:20250301\nDTEND;VALUE=DATE:20250332\nEND:VEVENT",
		},
		{
			// RFC 5545 §3.2.20: TZID MUST NOT be applied to DATE properties.
			// The DTEND case's VALUE=DATE branch rejects the combination
			// before parseICSDate is ever consulted. The DTSTART is a
			// well-formed VALUE=DATE so no earlier error preempts this one.
			name:  "TZID combined with VALUE=DATE on DTEND",
			input: "BEGIN:VEVENT\nDTSTART;VALUE=DATE:20250301\nDTEND;TZID=US-Eastern;VALUE=DATE:20250302\nEND:VEVENT",
		},
	}

	// Iterate over each test case and run it as a subtest.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new ICSScanner with the test input string.
			s := tsicsparser.NewICSScanner(strings.NewReader(tt.input), "test")
			// Consume the BEGIN:VEVENT line, mirroring how the calendar
			// parser hands the scanner to parseEvent after spotting BEGIN.
			s.Scan()
			// Call ParseEvent and verify that an error is returned.
			_, err := tsicsparser.ParseEvent(s, tt.tzs)
			if err == nil {
				t.Fatal(tserr.NilFailed(tt.name))
			}
		})
	}
}

// TestParseEventNestedBlockErr tests that a structural error inside a
// nested sub-component (e.g. VALARM) is surfaced by parseEvent through
// the collectRawBlock error branch in its case "BEGIN" arm, rather
// than being masked or misattributed to the outer VEVENT.
func TestParseEventNestedBlockErr(t *testing.T) {
	// A VEVENT containing a nested BEGIN:VALARM whose body is
	// malformed: END:VTODO appears at depth 0 inside the VALARM, so
	// collectRawBlock returns "unexpected END:VTODO inside VALARM".
	// parseEvent must propagate that error from its case "BEGIN" arm
	// (the collectRawBlock error branch), not from its own END handling
	// (which would say "inside VEVENT").
	input := "BEGIN:VEVENT\nDTSTART:20250101T120000Z\nBEGIN:VALARM\nEND:VTODO\nEND:VEVENT"
	// Create a new ICSScanner with the test input string.
	s := tsicsparser.NewICSScanner(strings.NewReader(input), "test")
	// Consume the BEGIN:VEVENT line, mirroring how the calendar parser
	// hands the scanner to parseEvent after spotting BEGIN.
	s.Scan()
	// Call ParseEvent and verify that an error is returned.
	_, err := tsicsparser.ParseEvent(s, nil)
	// If there is no error, the test has failed.
	if err == nil {
		t.Fatal(tserr.NilFailed("nested block error should surface"))
	}
	// The error must originate from collectRawBlock (referencing the
	// nested VALARM block), not from parseEvent's own END handling
	// (which would reference VEVENT). This proves the error propagated
	// through the collectRawBlock error branch.
	if !strings.Contains(err.Error(), "VALARM") {
		t.Fatal(tserr.UnexpectedError(&tserr.UnexpectedErrorArgs{
			Expected: errors.New("error referencing VALARM (from collectRawBlock)"),
			Actual:   err,
		}))
	}
}

// TestParseEventAllDay tests the VALUE=DATE handling in parseEvent:
// the date parses to midnight UTC, AllDay is set, and — with no DTEND
// or DURATION — End defaults to Start + 24h per RFC 5545 §3.6.1.
func TestParseEventAllDay(t *testing.T) {
	// A VEVENT with VALUE=DATE should parse to a midnight UTC start
	// time, and AllDay should be set.
	input := "BEGIN:VEVENT\nDTSTART;VALUE=DATE:20250301\nSUMMARY:All day\nEND:VEVENT"
	// Create a new ICSScanner with the test input string.
	s := tsicsparser.NewICSScanner(strings.NewReader(input), "test")
	// Consume the BEGIN:VEVENT line, mirroring how the calendar
	// parser hands the scanner to parseEvent after spotting BEGIN.
	s.Scan() // consume BEGIN:VEVENT, mirroring parseCalendar
	// Call ParseEvent and verify that an error is returned.
	ev, err := tsicsparser.ParseEvent(s, nil)
	// If there is an error parsing the event, we report it and stop the test.
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseEvent", Err: err}))
	}
	// Expected start is midnight UTC on March 1st.
	wantStart := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	// If event start does not match expected, report the mismatch.
	if !ev.Start.Equal(wantStart) {
		t.Error(tserr.EqualStr(&tserr.EqualStrArgs{
			Var: "start", Want: wantStart.Format(time.RFC3339), Actual: ev.Start.Format(time.RFC3339)}))
	}
	// If AllDay is not set, report the mismatch.
	if !ev.AllDay {
		t.Error(tserr.EqualStr(&tserr.EqualStrArgs{Var: "AllDay", Want: "true", Actual: "false"}))
	}
	// Expected end is start + 24h.
	wantEnd := wantStart.AddDate(0, 0, 1)
	// If event end does not match expected, report the mismatch.
	if !ev.End.Equal(wantEnd) {
		t.Error(tserr.EqualStr(&tserr.EqualStrArgs{
			Var: "end (+24h default)", Want: wantEnd.Format(time.RFC3339), Actual: ev.End.Format(time.RFC3339)}))
	}
}

// TestParseEventAllDayDTEND tests a VALUE=DATE DTSTART paired with a
// matching VALUE=DATE DTEND: the types match, so no error, and End is
// the exclusive end date — the +24h default is NOT applied because
// DTEND was given explicitly.
func TestParseEventAllDayDTEND(t *testing.T) {
	// A VEVENT with VALUE=DATE should parse to a midnight UTC start
	// time, and AllDay should be set.
	input := "BEGIN:VEVENT\nDTSTART;VALUE=DATE:20250301\nDTEND;VALUE=DATE:20250302\nEND:VEVENT"
	// Create a new ICSScanner with the test input string.
	s := tsicsparser.NewICSScanner(strings.NewReader(input), "test")
	// Consume the BEGIN:VEVENT line, mirroring how the calendar
	// parser hands the scanner to parseEvent after spotting BEGIN.
	s.Scan()
	// Call ParseEvent and verify that an error is returned.
	ev, err := tsicsparser.ParseEvent(s, nil)
	// If there is an error parsing the event, we report it and stop the test.
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseEvent", Err: err}))
	}
	// Expected start is midnight UTC on March 1st.
	wantStart := time.Date(2025, time.March, 1, 0, 0, 0, 0, time.UTC)
	// Expected end is midnight UTC on March 2nd.
	wantEnd := time.Date(2025, time.March, 2, 0, 0, 0, 0, time.UTC)
	// If event start does not match expected, report the mismatch.
	if !ev.Start.Equal(wantStart) {
		t.Error(tserr.EqualStr(&tserr.EqualStrArgs{Var: "start", Want: wantStart.Format(time.RFC3339), Actual: ev.Start.Format(time.RFC3339)}))
	}
	// If event end does not match expected, report the mismatch.
	if !ev.End.Equal(wantEnd) {
		t.Error(tserr.EqualStr(&tserr.EqualStrArgs{Var: "end", Want: wantEnd.Format(time.RFC3339), Actual: ev.End.Format(time.RFC3339)}))
	}
	// If event all-day flag does not match expected, report the mismatch.
	if !ev.AllDay {
		t.Error(tserr.EqualStr(&tserr.EqualStrArgs{Var: "allday", Want: "true", Actual: "false"}))
	}
}

// TestParseEventAllDayDuration tests day-aligned DURATIONs with a DATE
// DTSTART: P1D and P2W are valid per RFC 5545 §3.8.2.5, in both the
// DURATION-after-DTSTART and pending-DURATION-before-DTSTART orders.
func TestParseEventAllDayDuration(t *testing.T) {
	// Define a set of test cases
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{
			name:  "P1D after DTSTART",
			input: "BEGIN:VEVENT\nDTSTART;VALUE=DATE:20250301\nDURATION:P1D\nEND:VEVENT",
			want:  time.Date(2025, time.March, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "P2W after DTSTART",
			input: "BEGIN:VEVENT\nDTSTART;VALUE=DATE:20250301\nDURATION:P2W\nEND:VEVENT",
			want:  time.Date(2025, time.March, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "P1D before DTSTART (pending)",
			input: "BEGIN:VEVENT\nDURATION:P1D\nDTSTART;VALUE=DATE:20250301\nEND:VEVENT",
			want:  time.Date(2025, time.March, 2, 0, 0, 0, 0, time.UTC),
		},
	}
	// Iterate over each test case and run it as a subtest.
	for _, tt := range tests {
		// Run each test case as a subtest to isolate failures and provide better reporting.
		t.Run(tt.name, func(t *testing.T) {
			// Create a new ICSScanner with the test input string.
			s := tsicsparser.NewICSScanner(strings.NewReader(tt.input), "test")
			// Consume the BEGIN:VEVENT line, mirroring how the calendar
			// parser hands the scanner to parseEvent after spotting BEGIN.
			s.Scan()
			// Call ParseEvent and verify that an error is returned.
			ev, err := tsicsparser.ParseEvent(s, nil)
			// Report an error if there was an issue parsing the event.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseEvent", Fn: tt.name, Err: err}))
			}
			// Report an error if the actual end time does not match the expected time.
			if !ev.End.Equal(tt.want) {
				t.Error(tserr.EqualStr(&tserr.EqualStrArgs{
					Var: "end", Want: tt.want.Format(time.RFC3339), Actual: ev.End.Format(time.RFC3339)}))
			}
		})
	}
}
