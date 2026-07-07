// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsicsparser_test

import (
	"testing"

	"github.com/thorsphere/tserr"
	"github.com/thorsphere/tsfio"
	"github.com/thorsphere/tsicsparser"
)

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

	var err error

	for s.Scan() {
		line := s.Text()
		if line == "BEGIN:VTIMEZONE" {
			tz, err = tsicsparser.ParseTimezone(s)
			if err != nil {
				t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseTimezone", Fn: fn, Err: err}))
			}
			break
		}
	}

	// Check for any errors that occurred during scanning and report them.
	if err := s.Err(); err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Scan", Fn: fn, Err: err}))
	}

	tsfio.CreateGoldenFile(&tsfio.Testcase{Name: "timezone", Data: tz.String()})

	// Evaluate the parsed timezone against the golden file to ensure correctness.
	if err := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "timezone", Data: tz.String()}); err != nil {
		t.Fatal(err)
	}
}
