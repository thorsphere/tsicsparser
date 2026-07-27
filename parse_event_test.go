// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

import (
	"testing"
	"time"

	"github.com/thorsphere/tserr"
	"github.com/thorsphere/tsicsparser"
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
