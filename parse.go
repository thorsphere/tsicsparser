// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

// Import necessary packages for ICS parsing and error handling.
import "io" // Import the io package for I/O primitives.

// ParseCalendar reads ICS data from r and returns the parsed calendar,
// including its product identifier, timezones, and events.
// It handles line folding according to the ICS specification.
// Name is the name of the source calendar being scanned.
func Parse(r io.Reader, name string) (*Calendar, error) {
	// Create a new ICSScanner to read from the provided io.Reader.
	scanner := NewICSScanner(r, name)
	// Call the parseCalendar function to parse the calendar data from the scanner.
	return parseCalendar(scanner)
}
