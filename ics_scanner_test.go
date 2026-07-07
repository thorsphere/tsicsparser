// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsicsparser_test

// Import necessary packages for testing and ICS parsing.
import (
	"strings" // Import the strings package for string manipulation.
	"testing" // Import the testing package for writing test cases.

	"github.com/thorsphere/tserr"       // Import the tserr package for error handling.
	"github.com/thorsphere/tsfio"       // Import the tsfio package for file handling utilities.
	"github.com/thorsphere/tsicsparser" // Import the tsicsparser package to test the ICSScanner.
)

// TestICSScanner tests the ICSScanner by reading a sample ICS file and comparing the output
// against a golden file. It ensures that the scanner correctly handles line folding and
// produces the expected output.
func TestICSScanner(t *testing.T) {
	// Define the path to the ICS file that will be used for testing.
	fn := "testdata/scanner.ics"
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
	// Use a strings.Builder to efficiently build the output string from the scanned lines.
	b := strings.Builder{}
	// Scan through the ICS file, appending each line to the builder.
	for s.Scan() {
		b.WriteString(s.Text())
		b.WriteByte('\n')
	}
	// Check for any errors that occurred during scanning and report them.
	if err := s.Err(); err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "Scan", Fn: fn, Err: err}))
	}
	// Evaluate the output against the golden file to ensure correctness.
	if err := tsfio.EvalGoldenFile(&tsfio.Testcase{Name: "scanner", Data: b.String()}); err != nil {
		t.Fatal(err)
	}
}

// TestScanNil tests the behavior of the ICSScanner when it is nil.
// It ensures that calling Scan on a nil ICSScanner returns false, as expected.
func TestScanNil(t *testing.T) {
	// Test the behavior of the ICSScanner when it is nil.
	var s *tsicsparser.ICSScanner
	// Call Scan on a nil ICSScanner and check that it returns false.
	if s.Scan() {
		t.Fatal(tserr.NilFailed("Scan"))
	}
}

// TestTextNil tests the behavior of the ICSScanner when it is nil.
// It ensures that calling Text on a nil ICSScanner returns an empty string, as expected.
func TestTextNil(t *testing.T) {
	// Test the behavior of the ICSScanner when it is nil.
	var s *tsicsparser.ICSScanner
	// Call Text on a nil ICSScanner and check that it returns an empty string.
	if s.Text() != "" {
		t.Fatal(tserr.NilFailed("Text"))
	}
}

// TestErrNil tests the behavior of the ICSScanner when it is nil.
// It ensures that calling Err on a nil ICSScanner returns an error, as expected.
func TestErrNil(t *testing.T) {
	// Test the behavior of the ICSScanner when it is nil.
	var s *tsicsparser.ICSScanner
	// Call Err on a nil ICSScanner and check that it returns an error.
	if s.Err() == nil {
		t.Fatal(tserr.NilFailed("Err"))
	}
}
