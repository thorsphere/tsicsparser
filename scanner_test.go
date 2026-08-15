// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser_test

// Import necessary packages for testing and ICS parsing.
import (
	"errors"  // Import the errors package for error handling.
	"strings" // Import the strings package for string manipulation.
	"testing" // Import the testing package for writing test cases.

	"github.com/thorsphere/tserr"       // Import the tserr package for error handling.
	"github.com/thorsphere/tsfio"       // Import the tsfio package for file handling utilities.
	"github.com/thorsphere/tsicsparser" // Import the tsicsparser package to test the ICSScanner.
)

// faultyReader returns the given bytes on the first Read, then err on
// subsequent reads, simulating an I/O failure mid-stream.
type faultyReader struct {
	data []byte // The data to be read.
	err  error  // The error to be returned on subsequent reads.
	pos  int    // The current position in the data.
}

// Read implements the io.Reader interface for faultyReader.
// first Read, then err on subsequent reads, simulating an I/O failure
// mid-stream.
func (r *faultyReader) Read(p []byte) (int, error) {
	// If we have reached the end of the data, return the error.
	if r.pos >= len(r.data) {
		return 0, r.err
	}
	// Copy the data from the current position to the end of the buffer.
	n := copy(p, r.data[r.pos:])
	// Advance the position by the number of bytes read.
	r.pos += n
	// Return the number of bytes read and nil to indicate success.
	return n, nil
}

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

// TestParseCalScannerErr tests that an I/O error from the
// scanner is surfaced, not masked by "Unexpected end of input".
func TestParseCalScannerErr(t *testing.T) {
	// A valid calendar header up to the point where the reader fails.
	head := "BEGIN:VCALENDAR\nVERSION:2.0\nPRODID:+//Thorsphere//tsicsparser//en\n"
	// Create a new ICSScanner to read from a faulty reader.
	s := tsicsparser.NewICSScanner(
		&faultyReader{data: []byte(head), err: errors.New("disk read failed")},
		"test")
	// Call ParseCalendar on the faulty scanner and check that it
	// returns an error.
	_, err := tsicsparser.ParseCalendar(s)
	// If there is no error, the test has failed.
	if err == nil {
		t.Fatal(tserr.NilFailed("scanner I/O error should surface"))
	}
	// The error must be the I/O error, not the structural
	// "Unexpected end of input while parsing calendar".
	if !strings.Contains(err.Error(), "disk read failed") {
		t.Fatalf("expected scanner I/O error, got: %v", err)
	}
}
