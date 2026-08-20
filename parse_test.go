// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import necessary packages for testing the tsicsparser package.
import (
	"testing" // Import the testing package to write unit tests for the tsicsparser package.

	"github.com/thorsphere/tserr"       // Import the tserr package to handle error reporting and formatting in the tests.
	"github.com/thorsphere/tsfio"       // Import the tsfio package to handle file input/output operations in the tests.
	"github.com/thorsphere/tsicsparser" // Import the tsicsparser package to test the Parse function.
)

// TestParse tests the exported Parse function by reading the sample
// calendar from testdata/cal.ics and comparing the parsed output
// against the golden file. It exercises the full public pipeline —
// reader, BOM stripping, line folding, two-pass event/timezone
// resolution — and must produce output identical to the internal
// parseCalendar path tested by TestParseCal.
func TestParse(t *testing.T) {
	// Define the path to the ICS file that will be used for testing.
	fn := "testdata/cal.ics"
	// Open the ICS file for reading using the tsfio package, which provides file handling utilities.
	f, e := tsfio.OpenFile(tsfio.Filename(fn))
	// If there is an error opening the file, we report an error and stop the test.
	if e != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "OpenFile", Fn: fn, Err: e}))
	}
	// Ensure the file is closed after the test completes to avoid resource leaks.
	defer f.Close()
	// Call the Parse function to parse the calendar data directly from the file reader.
	cal, err := tsicsparser.Parse(f, fn)
	// If there is an error while parsing the calendar, we report an error and stop the test.
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Parse", Fn: fn, Err: err}))
	}
	// Evaluate the parsed calendar against the golden file to ensure correctness.
	// This reuses testdata/cal.golden: Parse must yield the same result as the
	// scanner-based path, so the same golden file is the reference.
	if err := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "cal", Data: cal.String()}); err != nil {
		t.Fatal(err)
	}
}
