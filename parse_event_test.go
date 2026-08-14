// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import the necessary packages for testing, string manipulation, and custom error handling.
import (
	"fmt"     // For formatting error messages.
	"strings" // For string manipulation.
	"testing" // For writing test cases.
	"time"

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

// TestParseEventErrors tests the parseEvent function with various malformed ICS
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
