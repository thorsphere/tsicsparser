// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import necessary packages for testing and ICS parsing.
import (
	"strings" // For string manipulation.
	"testing" // For writing test cases.

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
