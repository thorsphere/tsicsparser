// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsicsparser_test

// Import necessary packages for testing and ICS parsing.
import (
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
