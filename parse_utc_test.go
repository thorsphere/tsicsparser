// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import necessary packages for testing and ICS parsing.
import (
	"fmt"     // For formatting error messages.
	"testing" // For writing test cases.
	"time"    // For time operations.

	"github.com/thorsphere/lpstats"     // For comparing maps of strings.
	"github.com/thorsphere/tserr"       // For error handling.
	"github.com/thorsphere/tsicsparser" // For testing the tsicsparser package.
)

// TestParseWeekday tests the ParseWeekday function.
// It checks if the function correctly parses the weekday from a string.
// It fails if the function returns an error.
func TestParseWeekday(t *testing.T) {
	// Define a test case for each weekday.
	tests := []struct {
		input string
		want  time.Weekday
	}{
		{"SU", time.Sunday},
		{"MO", time.Monday},
		{"TU", time.Tuesday},
		{"WE", time.Wednesday},
		{"TH", time.Thursday},
		{"FR", time.Friday},
		{"SA", time.Saturday},
	}
	// Loop through the tests and check if the weekday matches the expected weekday.
	for _, tt := range tests {
		// Parse the weekday from the input string.
		wd, err := tsicsparser.ParseWeekday(tt.input)
		// Check if there is an error parsing the weekday.
		if err != nil {
			t.Errorf("ParseWeekday(%s) error: %v", tt.input, err)
		}
		// Check if the weekday matches the expected weekday.
		if wd != tt.want {
			t.Error(tserr.EqualInt(&tserr.EqualIntArgs{Var: "weekday", Want: int64(tt.want), Actual: int64(wd)}))
		}
	}
}

// TestParseWeekdayErr tests the ParseWeekday function.
// It checks if the function returns an error for an invalid weekday. it fails if the function returns nil.
func TestParseWeekdayErr(t *testing.T) {
	// Check if the function returns an error for an invalid weekday.
	_, err := tsicsparser.ParseWeekday("invalid")
	// Check if the error is not nil.
	if err == nil {
		t.Error(tserr.NilFailed("ParseWeekday"))
	}
}

// TestParseByDay tests the ParseByDay function.
// It checks if the function correctly parses the ordinal and weekday from a string.
// It fails if the function returns an error.
func TestParseByDay(t *testing.T) {
	// Define a test case for each weekday.
	tests := []struct {
		input string
		or    int
		wd    time.Weekday
	}{
		{"-1MO", -1, time.Monday},
		{"1TU", 1, time.Tuesday},
		{"-2WE", -2, time.Wednesday},
		{"2TH", 2, time.Thursday},
		{"-3FR", -3, time.Friday},
		{"3SA", 3, time.Saturday},
		{"SU", 0, time.Sunday},
	}
	// Loop through the tests and check if the ordinal and weekday match the expected values.
	for _, tt := range tests {
		// Parse the weekday from the input string.
		ordinal, weekday, err := tsicsparser.ParseByDay(tt.input)
		// Check if there is an error parsing the weekday.
		if err != nil {
			t.Error(tserr.NilExpected("parseByDay"))
		}
		// Check if the ordinal matches the expected ordinal.
		if ordinal != tt.or {
			t.Error(tserr.EqualInt(&tserr.EqualIntArgs{Var: "ordinal", Want: int64(tt.or), Actual: int64(ordinal)}))
		}
		// Check if the weekday matches the expected weekday.
		if weekday != tt.wd {
			t.Error(tserr.EqualInt(&tserr.EqualIntArgs{Var: "weekday", Want: int64(tt.wd), Actual: int64(weekday)}))
		}
	}
}

// TestParseByDayErr tests the ParseByDay function.
// It checks if the function returns an error for an invalid weekday. it fails if the function returns nil.
func TestParseByDayErr(t *testing.T) {
	// Define a test case for each invalid weekday.
	tests := []struct {
		input string
	}{
		{"-1INVALID"},
		{"1IN"},
		{"iMO"},
		{"M"},
	}
	// Loop through the tests and check if the function returns an error.
	for _, tt := range tests {
		// Parse the weekday from the input string.
		_, _, err := tsicsparser.ParseByDay(tt.input)
		// Check if there is an error parsing the weekday.
		if err == nil {
			t.Error(tserr.NilFailed("parseByDay"))
		}
	}
}

// TestNthWeekday tests the NthWeekday function.
// It checks if the function correctly parses the ordinal and weekday from a string.
// It fails if the function returns an error.
func TestNthWeekday(t *testing.T) {
	// Define a test case for each weekday.
	tests := []struct {
		year  int
		month time.Month
		byDay string
		want  int
	}{
		{1984, time.June, "2SU", 10},
		{1984, time.June, "-2SU", 17},
	}
	// Loop through the tests and check if the weekday matches the expected weekday.
	for _, tt := range tests {
		// Parse the weekday from the input string.
		wd, err := tsicsparser.NthWeekday(tt.year, tt.month, tt.byDay)
		// Check if there is an error parsing the weekday.
		if err != nil {
			t.Fatal(tserr.Op(&tserr.OpArgs{Op: "nthWeekday", Err: err}))
		}
		// Check if the weekday matches the expected weekday.
		// If the correct weekday is not recognized, it returns an error.
		if wd != tt.want {
			t.Error(tserr.EqualInt(&tserr.EqualIntArgs{Var: "weekday", Actual: int64(wd), Want: int64(tt.want)}))
		}
	}
}

// TestNthWeekdayErr tests the NthWeekday function.
// It checks if the function returns an error for an invalid weekday.
// It fails if the function returns nil.
func TestNthWeekdayErr(t *testing.T) {
	// Define a test case for each invalid weekday.
	tests := []struct {
		year  int
		month time.Month
		byDay string
	}{
		{1984, time.June, "-1INVALID"},
		{1984, time.June, "1IN"},
		{1984, time.June, "iMO"},
		{1984, time.June, "MO"},
		{1984, time.June, "5MO"},
		{1984, time.June, "-5MO"},
	}
	// Loop through the tests and check if the function returns an error.
	for _, tt := range tests {
		// Parse the weekday from the input string.
		_, err := tsicsparser.NthWeekday(tt.year, tt.month, tt.byDay)
		// Check if there is an error parsing the weekday.
		if err == nil {
			t.Error(tserr.NilFailed("nthWeekday"))
		}
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

// TestTransitionTime tests the TransitionTime function.
// It checks that the function correctly computes the local datetime when
// a timezone rule takes effect, using either the rule's RRULE (recurring)
// or the rule's DTSTART (fixed). It fails if the function returns an error
// or if the computed transition time doesn't match the expected value.
func TestTransitionTime(t *testing.T) {
	// Define a set of test cases covering both recurring (RRULE-based) and
	// fixed (DTSTART-only) timezone transitions.
	tests := []struct {
		name string
		rule tsicsparser.TimezoneRule
		year int
		want time.Time
	}{
		{
			// US Eastern: daylight saving begins on the 2nd Sunday of March at 02:00.
			// In 2024, the 2nd Sunday of March is March 10.
			name: "recurring daylight US-Eastern 2024",
			rule: tsicsparser.TimezoneRule{
				Type:    tsicsparser.Daylight,
				DTStart: time.Date(0, time.March, 10, 2, 0, 0, 0, time.UTC),
				RRule: &tsicsparser.RRule{
					Freq:    "YEARLY",
					ByMonth: int(time.March),
					ByDay:   "2SU",
				},
			},
			year: 2024,
			want: time.Date(2024, time.March, 10, 2, 0, 0, 0, time.UTC),
		},
		{
			// US Eastern: daylight saving begins on the 2nd Sunday of March at 02:00.
			// In 2024, the 2nd Sunday of March is March 10.
			// The RRULE's ByMonth is missing (0), so the function falls back to DTStart's month (March).
			name: "recurring daylight US-Eastern 2024 fallback when ByMonth is 0",
			rule: tsicsparser.TimezoneRule{
				Type:    tsicsparser.Daylight,
				DTStart: time.Date(0, time.March, 10, 2, 0, 0, 0, time.UTC),
				RRule: &tsicsparser.RRule{
					Freq:    "YEARLY",
					ByMonth: 0,
					ByDay:   "2SU",
				},
			},
			year: 2024,
			want: time.Date(2024, time.March, 10, 2, 0, 0, 0, time.UTC),
		},
		{
			// US Eastern: daylight saving begins on the 2nd Sunday of March at 02:00.
			// In 2025, the 2nd Sunday of March is March 9.
			name: "recurring daylight US-Eastern 2025",
			rule: tsicsparser.TimezoneRule{
				Type:    tsicsparser.Daylight,
				DTStart: time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
				RRule: &tsicsparser.RRule{
					Freq:    "YEARLY",
					ByMonth: int(time.March),
					ByDay:   "2SU",
				},
			},
			year: 2025,
			want: time.Date(2025, time.March, 9, 2, 0, 0, 0, time.UTC),
		},
		{
			// US Eastern: standard time begins on the 1st Sunday of November at 02:00.
			// In 2024, the 1st Sunday of November is November 3.
			name: "recurring standard US-Eastern 2024",
			rule: tsicsparser.TimezoneRule{
				Type:    tsicsparser.Standard,
				DTStart: time.Date(0, time.November, 3, 2, 0, 0, 0, time.UTC),
				RRule: &tsicsparser.RRule{
					Freq:    "YEARLY",
					ByMonth: int(time.November),
					ByDay:   "1SU",
				},
			},
			year: 2024,
			want: time.Date(2024, time.November, 3, 2, 0, 0, 0, time.UTC),
		},
		{
			// Last Sunday of October at 03:00.
			// In 2024, the last Sunday of October is October 27.
			name: "recurring last Sunday October 2024",
			rule: tsicsparser.TimezoneRule{
				Type:    tsicsparser.Daylight,
				DTStart: time.Date(0, time.October, 27, 3, 0, 0, 0, time.UTC),
				RRule: &tsicsparser.RRule{
					Freq:    "YEARLY",
					ByMonth: int(time.October),
					ByDay:   "-1SU",
				},
			},
			year: 2024,
			want: time.Date(2024, time.October, 27, 3, 0, 0, 0, time.UTC),
		},
		{
			// Fixed transition: no RRULE, the rule returns DTSTART as-is.
			// This is the legacy fallback for timezones that pre-date RRULE usage.
			name: "fixed transition without RRule",
			rule: tsicsparser.TimezoneRule{
				Type:    tsicsparser.Standard,
				DTStart: time.Date(2007, time.March, 11, 2, 0, 0, 0, time.UTC),
			},
			year: 2024,
			want: time.Date(2024, time.March, 11, 2, 0, 0, 0, time.UTC),
		},
		{
			// Time-of-day preservation: only the date should change between years;
			// the hour/minute/second come from DTSTART.
			name: "time of day preserved from DTSTART",
			rule: tsicsparser.TimezoneRule{
				Type:    tsicsparser.Daylight,
				DTStart: time.Date(0, time.March, 25, 1, 30, 45, 0, time.UTC),
				RRule: &tsicsparser.RRule{
					Freq:    "YEARLY",
					ByMonth: int(time.March),
					ByDay:   "4SU",
				},
			},
			year: 2030,
			want: time.Date(2030, time.March, 24, 1, 30, 45, 0, time.UTC),
		},
	}
	// Run each test case as a subtest for clearer failure attribution.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compute the transition time for the given rule and year.
			got, err := tsicsparser.TransitionTime(tt.rule, tt.year)
			// Fail immediately if an unexpected error is returned.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "transitionTime", Err: err}))
			}
			// Check that the computed transition time matches the expected value.
			if !got.Equal(tt.want) {
				t.Errorf("transitionTime(year=%d) = %s, want %s",
					tt.year, got.Format(time.RFC3339), tt.want.Format(time.RFC3339))
			}
		})
	}
}

// TestConvertToUTC tests the ConvertToUTC function across all of its
// branches: Northern Hemisphere DST, Southern Hemisphere DST, single-rule
// timezones (standard-only and daylight-only), empty TZID, TZID mismatch,
// and the no-rules / no-daylight-no-standard error paths. It fails if the
// function returns an unexpected error or returns a time that doesn't
// match the expected UTC value.
func TestConvertToUTC(t *testing.T) {
	// Define a set of test cases covering the full matrix of convertToUTC behavior.
	tests := []struct {
		name      string
		localTime time.Time
		tzid      string
		tz        tsicsparser.Timezone
		want      time.Time
	}{
		{
			// Northern Hemisphere: March 9 2025 is after the 2nd-Sunday-of-March
			// transition (Mar 9 at 02:00 local) but before the 1st-Sunday-of-November
			// transition (Nov 2 at 02:00 local). Daylight (UTC-4) is in effect.
			// 12:00 local → 16:00 UTC.
			name:      "northern hemisphere summer daylight",
			localTime: time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "US-Eastern",
			tz: tsicsparser.Timezone{
				TZID: "US-Eastern",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: -5 * 3600,
						TZOffsetTo:   -4 * 3600,
						DTStart:      time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: -4 * 3600,
						TZOffsetTo:   -5 * 3600,
						DTStart:      time.Date(0, time.November, 2, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.June, 15, 16, 0, 0, 0, time.UTC),
		},
		{
			// Northern Hemisphere: January 15 is before the spring transition.
			// Standard (UTC-5) is in effect. 12:00 local → 17:00 UTC.
			name:      "northern hemisphere winter standard",
			localTime: time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "US-Eastern",
			tz: tsicsparser.Timezone{
				TZID: "US-Eastern",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: -5 * 3600,
						TZOffsetTo:   -4 * 3600,
						DTStart:      time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: -4 * 3600,
						TZOffsetTo:   -5 * 3600,
						DTStart:      time.Date(0, time.November, 2, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.January, 15, 17, 0, 0, 0, time.UTC),
		},
		{
			// Northern Hemisphere: on the boundary date Nov 2 2025 at 01:30 local
			// — before the 02:00 transition that ends DST. Daylight (UTC-4) still active.
			// 01:30 local → 05:30 UTC.
			name:      "northern hemisphere before fall transition",
			localTime: time.Date(2025, time.November, 2, 1, 30, 0, 0, time.UTC),
			tzid:      "US-Eastern",
			tz: tsicsparser.Timezone{
				TZID: "US-Eastern",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: -5 * 3600,
						TZOffsetTo:   -4 * 3600,
						DTStart:      time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: -4 * 3600,
						TZOffsetTo:   -5 * 3600,
						DTStart:      time.Date(0, time.November, 2, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.November, 2, 5, 30, 0, 0, time.UTC),
		},
		{
			// Northern Hemisphere: on Nov 2 2025 at 03:00 local — after the 02:00 transition.
			// Standard (UTC-5) is now in effect. 03:00 local → 08:00 UTC.
			name:      "northern hemisphere after fall transition",
			localTime: time.Date(2025, time.November, 2, 3, 0, 0, 0, time.UTC),
			tzid:      "US-Eastern",
			tz: tsicsparser.Timezone{
				TZID: "US-Eastern",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: -5 * 3600,
						TZOffsetTo:   -4 * 3600,
						DTStart:      time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: -4 * 3600,
						TZOffsetTo:   -5 * 3600,
						DTStart:      time.Date(0, time.November, 2, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.November, 2, 8, 0, 0, 0, time.UTC),
		},
		{
			// Northern Hemisphere: on Nov 2 2025 at 02:00 local — at the 02:00 transition.
			// Standard (UTC-5) is now in effect. 02:00 local → 07:00 UTC.
			name:      "northern hemisphere at fall transition",
			localTime: time.Date(2025, time.November, 2, 2, 0, 0, 0, time.UTC),
			tzid:      "US-Eastern",
			tz: tsicsparser.Timezone{
				TZID: "US-Eastern",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: -5 * 3600,
						TZOffsetTo:   -4 * 3600,
						DTStart:      time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: -4 * 3600,
						TZOffsetTo:   -5 * 3600,
						DTStart:      time.Date(0, time.November, 2, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.November, 2, 7, 0, 0, 0, time.UTC),
		},
		{
			// Northern Hemisphere: on Mar 9 2025 at 02:00 local — at the 02:00 transition.
			// Daylight (UTC-4) is now in effect. 02:00 local → 06:00 UTC.
			name:      "northern hemisphere at spring transition",
			localTime: time.Date(2025, time.March, 9, 2, 0, 0, 0, time.UTC),
			tzid:      "US-Eastern",
			tz: tsicsparser.Timezone{
				TZID: "US-Eastern",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: -5 * 3600,
						TZOffsetTo:   -4 * 3600,
						DTStart:      time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: -4 * 3600,
						TZOffsetTo:   -5 * 3600,
						DTStart:      time.Date(0, time.November, 2, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.March, 9, 6, 0, 0, 0, time.UTC),
		},
		{
			// Southern Hemisphere: Sydney. Daylight starts in October, standard in April.
			// Jan 15 2025 is in the middle of daylight (UTC+11). 12:00 local → 01:00 UTC.
			name:      "southern hemisphere summer daylight",
			localTime: time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "Australia/Sydney",
			tz: tsicsparser.Timezone{
				TZID: "Australia/Sydney",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: 10 * 3600,
						TZOffsetTo:   11 * 3600,
						DTStart:      time.Date(0, time.October, 6, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.October), ByDay: "1SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: 11 * 3600,
						TZOffsetTo:   10 * 3600,
						DTStart:      time.Date(0, time.April, 6, 3, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.April), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.January, 15, 1, 0, 0, 0, time.UTC),
		},
		{
			// Southern Hemisphere: July 15 2025 — between the April standard transition and
			// the October daylight transition.
			// Standard (UTC+10) is in effect. 12:00 local → 02:00 UTC.
			name:      "southern hemisphere winter standard",
			localTime: time.Date(2025, time.July, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "Australia/Sydney",
			tz: tsicsparser.Timezone{
				TZID: "Australia/Sydney",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: 10 * 3600,
						TZOffsetTo:   11 * 3600,
						DTStart:      time.Date(0, time.October, 6, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.October), ByDay: "1SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: 11 * 3600,
						TZOffsetTo:   10 * 3600,
						DTStart:      time.Date(0, time.April, 6, 3, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.April), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.July, 15, 2, 0, 0, 0, time.UTC),
		},
		{
			// Southern Hemisphere: October 6 2025 — at the October transition to spring.
			// Daylight (UTC+11) is in effect. 2:00 local → 15:00 UTC.
			name:      "southern hemisphere at spring transition",
			localTime: time.Date(2025, time.October, 6, 2, 0, 0, 0, time.UTC),
			tzid:      "Australia/Sydney",
			tz: tsicsparser.Timezone{
				TZID: "Australia/Sydney",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: 10 * 3600,
						TZOffsetTo:   11 * 3600,
						DTStart:      time.Date(0, time.October, 6, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.October), ByDay: "1SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: 11 * 3600,
						TZOffsetTo:   10 * 3600,
						DTStart:      time.Date(0, time.April, 6, 3, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.April), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.October, 5, 15, 0, 0, 0, time.UTC),
		},
		{
			// Southern Hemisphere: April 6 2025 — at the April standard transition (SH fall back).
			// At exactly the transition, standard (UTC+10) is in effect. 03:00 local → 17:00 UTC.
			name:      "southern hemisphere at fall transition",
			localTime: time.Date(2025, time.April, 6, 3, 0, 0, 0, time.UTC),
			tzid:      "Australia/Sydney",
			tz: tsicsparser.Timezone{
				TZID: "Australia/Sydney",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: 10 * 3600,
						TZOffsetTo:   11 * 3600,
						DTStart:      time.Date(0, time.October, 6, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.October), ByDay: "1SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: 11 * 3600,
						TZOffsetTo:   10 * 3600,
						DTStart:      time.Date(0, time.April, 6, 3, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.April), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.April, 5, 17, 0, 0, 0, time.UTC),
		},
		{
			// Single-rule timezone (Standard only, no DST). UTC+0 no transitions.
			// 12:00 local → 12:00 UTC.
			name:      "standard-only no DST",
			localTime: time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "UTC",
			tz: tsicsparser.Timezone{
				TZID: "UTC",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:       tsicsparser.Standard,
						TZOffsetTo: 0,
						DTStart:    time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC),
		},
		{
			// Single-rule timezone (Daylight only). The TZOffsetTo is applied directly.
			// Represents a fixed-UTC+5:30 zone (e.g., legacy India without a STANDARD rule).
			// 12:00 local → 06:30 UTC.
			name:      "daylight-only no standard rule",
			localTime: time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "Custom-IST",
			tz: tsicsparser.Timezone{
				TZID: "Custom-IST",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:       tsicsparser.Daylight,
						TZOffsetTo: 5*3600 + 30*60,
						DTStart:    time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: time.Date(2025, time.June, 15, 6, 30, 0, 0, time.UTC),
		},
		{
			// Edge case: negative TZOffsetTo. Timezone is east of UTC.
			// Berlin standard is UTC+1, so 12:00 local → 11:00 UTC.
			name:      "positive offset standard-only Europe/Berlin winter",
			localTime: time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "Europe/Berlin",
			tz: tsicsparser.Timezone{
				TZID: "Europe/Berlin",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:       tsicsparser.Standard,
						TZOffsetTo: 1 * 3600,
						DTStart:    time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			want: time.Date(2025, time.January, 15, 11, 0, 0, 0, time.UTC),
		},
		{
			// Northern Hemisphere: after the spring-forward boundary at 02:30 local.
			// Per the convertToUTC logic, after the transition time, the new
			// (daylight) offset is in effect. 02:30 local → 06:30 UTC (UTC-4).
			name:      "northern hemisphere after the spring-forward boundary at 02:30 local",
			localTime: time.Date(2025, time.March, 9, 2, 30, 0, 0, time.UTC),
			tzid:      "US-Eastern",
			tz: tsicsparser.Timezone{
				TZID: "US-Eastern",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: -5 * 3600,
						TZOffsetTo:   -4 * 3600,
						DTStart:      time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
					},
					{
						Type:         tsicsparser.Standard,
						TZOffsetFrom: -4 * 3600,
						TZOffsetTo:   -5 * 3600,
						DTStart:      time.Date(0, time.November, 2, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
					},
				},
			},
			want: time.Date(2025, time.March, 9, 6, 30, 0, 0, time.UTC),
		},
	}
	// Run each test case as a subtest for clearer failure attribution.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert the local time to UTC using the given timezone.
			got, err := tsicsparser.ConvertToUTC(tt.localTime, tt.tzid, tt.tz)
			// Fail immediately if an unexpected error is returned.
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "convertToUTC", Err: err}))
			}
			// Check that the converted time matches the expected UTC value.
			if !got.Equal(tt.want) {
				t.Errorf("convertToUTC(%s, %s) = %s, want %s",
					tt.localTime.Format(time.RFC3339),
					tt.tzid,
					got.Format(time.RFC3339),
					tt.want.Format(time.RFC3339))
			}
		})
	}
}

// TestConvertToUTCErr tests the ConvertToUTC function.
// It checks that the function returns an error for invalid inputs:
// empty TZID, TZID mismatch, timezones with no rules, timezones with
// neither daylight nor standard rules, and rules whose RRULE contains
// an invalid BYDAY. It fails if the function returns nil for any case.
func TestConvertToUTCErr(t *testing.T) {
	// Define a reusable "valid" US-Eastern ruleset for the BYDAY-error case.
	validRules := []tsicsparser.TimezoneRule{
		{
			Type:         tsicsparser.Daylight,
			TZOffsetFrom: -5 * 3600,
			TZOffsetTo:   -4 * 3600,
			DTStart:      time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
			RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "2SU"},
		},
		{
			Type:         tsicsparser.Standard,
			TZOffsetFrom: -4 * 3600,
			TZOffsetTo:   -5 * 3600,
			DTStart:      time.Date(0, time.November, 2, 2, 0, 0, 0, time.UTC),
			RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.November), ByDay: "1SU"},
		},
	}
	// Define a set of test cases that should all return an error.
	tests := []struct {
		name      string
		localTime time.Time
		tzid      string
		tz        tsicsparser.Timezone
	}{
		{
			name:      "empty TZID",
			localTime: time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "",
			tz: tsicsparser.Timezone{
				TZID:  "US-Eastern",
				Rules: validRules,
			},
		},
		{
			name:      "TZID mismatch",
			localTime: time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "Europe/London",
			tz: tsicsparser.Timezone{
				TZID:  "US-Eastern",
				Rules: validRules,
			},
		},
		{
			name:      "no rules",
			localTime: time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "Empty",
			tz: tsicsparser.Timezone{
				TZID:  "Empty",
				Rules: nil,
			},
		},
		{
			name:      "no daylight and no standard",
			localTime: time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "Weird",
			tz: tsicsparser.Timezone{
				TZID: "Weird",
				Rules: []tsicsparser.TimezoneRule{
					// A rule with an unrecognized type — neither Daylight nor Standard.
					{Type: tsicsparser.RuleType(99), DTStart: time.Now()},
				},
			},
		},
		{
			name:      "invalid BYDAY in RRULE",
			localTime: time.Date(2025, time.June, 15, 12, 0, 0, 0, time.UTC),
			tzid:      "US-Eastern",
			tz: tsicsparser.Timezone{
				TZID: "US-Eastern",
				Rules: []tsicsparser.TimezoneRule{
					{
						Type:         tsicsparser.Daylight,
						TZOffsetFrom: -5 * 3600,
						TZOffsetTo:   -4 * 3600,
						DTStart:      time.Date(0, time.March, 9, 2, 0, 0, 0, time.UTC),
						RRule:        &tsicsparser.RRule{Freq: "YEARLY", ByMonth: int(time.March), ByDay: "INVALID"},
					},
					validRules[1],
				},
			},
		},
	}
	// Loop through the tests and verify that each one returns an error.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt to convert the local time to UTC.
			_, err := tsicsparser.ConvertToUTC(tt.localTime, tt.tzid, tt.tz)
			// Verify that an error is returned for the invalid input.
			if err == nil {
				t.Error(tserr.NilFailed("convertToUTC"))
			}
		})
	}
}
