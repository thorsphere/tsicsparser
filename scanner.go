// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

// Import necessary packages for ICS parsing and error handling.
import (
	"bufio"   // Import the bufio package for buffered I/O operations.
	"io"      // Import the io package for I/O primitives.
	"strings" // Import the strings package for string manipulation.

	"github.com/thorsphere/tserr" // Import the tserr package for error handling.
)

// ICSScanner is a scanner that reads lines from an ICS file,
// handling line folding according to the ICS specification.
type ICSScanner struct {
	scanner     *bufio.Scanner // The underlying bufio.Scanner that reads from the input stream.
	name        string         // The name of the ICS file or source being scanned.
	currentLine string         // The current line that has been scanned and is returned by Text().
	nextLine    string         // The next line that has been scanned and is ready to be processed.
	hasLooked   bool           // A flag indicating whether we have looked ahead to the next line.
	currentErr  error          // Any error that occurred during scanning.
}

// NewICSScanner creates a new ICSScanner that reads from the provided io.Reader.
func NewICSScanner(r io.Reader, name string) *ICSScanner {
	// Return a new ICSScanner that reads from the provided io.Reader.
	// The scanner will handle line folding according to the ICS specification.
	scanner := bufio.NewScanner(r)
	// Allocate a larger buffer (e.g., 1MB) to handle large attachments or descriptions
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max capacity
	// Return a new ICSScanner that reads from the provided io.Reader.
	return &ICSScanner{
		scanner: scanner,
		name:    name,
	}
}

// Scan reads the next line from the input stream, handling line folding according
// to the ICS specification. Lines of text SHOULD NOT be longer than 75 octets,
// excluding the line break. Long content lines SHOULD be split into a multiple line
// representations using a line "folding" technique. That is, a long line can be split
// between any two characters by inserting a CRLF immediately followed by a single
// linear white-space character (i.e., SPACE or HTAB). Any sequence of CRLF followed
// immediately by a single linear white-space character is ignored (i.e., removed)
// when processing the content type.
func (s *ICSScanner) Scan() bool {
	// If the scanner is nil, we cannot continue scanning, so we return false.
	if s == nil {
		return false
	}
	// If the current error is not nil, we cannot continue scanning, so we return false.
	if s.currentErr != nil {
		// If there is an error, we cannot continue scanning, so we return false.
		return false
	}
	// Read the next line from the scanner, handling lookahead and
	// returning the line and a boolean indicating success.
	line, ok := s.readLine()
	// If we cannot read the next line, we return false to indicate that we cannot continue scanning.
	if !ok {
		return false
	}
	// Peek at the next line to see if it is a folded line (i.e., starts with a space or tab).
	peek, ok := s.readLine()
	// If the next line is a folded line, we need to combine it with the current line.
	if ok && len(peek) > 0 && (peek[0] == ' ' || peek[0] == '\t') {
		// String builder to efficiently concatenate the current line with the folded lines.
		var b strings.Builder
		// Write the current line to the string builder.
		b.WriteString(line)
		// Write the next line to the string builder, skipping the leading space/tab.
		b.WriteString(peek[1:])
		// Continue reading and combining folded lines until we reach a line that is not folded.
		for {
			// Peek at the next line to see if it is a folded line (i.e., starts with a space or tab).
			peek, ok = s.readLine()
			// If the next line is not a folded line, we need to store it for the next Scan() call.
			if !ok || len(peek) == 0 || (peek[0] != ' ' && peek[0] != '\t') {
				// If ok is true, we have a valid next line that is not folded,
				// so we store it for the next Scan() call.
				if ok {
					// Store it for the next Scan() call.
					s.nextLine = peek
					// Set hasLooked to true to indicate that
					// we have looked ahead and have a next line.
					s.hasLooked = true
				}
				// Set the current line to the combined line
				s.currentLine = b.String()
				// Return true to indicate that we have successfully scanned a line.
				return true
			}
			// Combine it with the current line.
			b.WriteString(peek[1:])
		}
	}
	// If we have a valid next line that is not folded, we need to store it for the next Scan() call.
	if ok {
		s.nextLine = peek
		// Set hasLooked to true to indicate that we have looked ahead and have a next line.
		s.hasLooked = true
	}
	// Set the current line to the line we just read.
	s.currentLine = line
	// Return true to indicate that we have successfully scanned a line.
	return true
}

// Text returns the current line of text that has been scanned and combined, including any folded lines.
func (s *ICSScanner) Text() string {
	// If the scanner is nil, we cannot return any text, so we return an empty string.
	if s == nil {
		return ""
	}
	// Return the current line of text that has been scanned and combined, including any folded lines.
	return s.currentLine
}

// Err returns any I/O errors that occurred during scanning.
func (s *ICSScanner) Err() error {
	// If the scanner is nil, we cannot return any errors, so we return nil.
	if s == nil {
		return tserr.NilPtr()
	}
	// Return the current error that occurred during scanning, if any.
	return s.currentErr
}

// readLine reads the next line from the scanner, handling lookahead
// and returning the line and a boolean indicating success.
func (s *ICSScanner) readLine() (string, bool) {
	// If the scanner is nil, we cannot read any lines, so we return an empty string and false.
	if s == nil {
		return "", false
	}
	// If we have already looked ahead and have a next line, we return that line.
	if s.hasLooked {
		// If we have already looked ahead and have a next line, we return that line.
		s.hasLooked = false
		return s.nextLine, true
	}
	// If we haven't looked ahead, we read the next line from the scanner.
	if s.scanner.Scan() {
		// If we can read the next line, we return the cleaned and unescaped line
		// and true to indicate success.
		return unescapeText(cleanLine(s.scanner.Text())), true
	}
	// If we cannot read the next line, we check if there was an error during scanning.
	if err := s.scanner.Err(); err != nil {
		// If there is an error, we return the error.
		s.currentErr = err
	}
	// If we cannot read the next line, we return an empty string and
	// false to indicate that we cannot continue scanning.
	return "", false
}

// cleanLine removes trailing carriage returns (\r) from the end of a line.
func cleanLine(line string) string {
	// Remove trailing carriage returns (\r) from the end of the line,
	// as they are not part of the actual content.
	return strings.TrimRight(line, "\r")
}

// unescapeText removes RFC 5545 escape characters from text fields
func unescapeText(input string) string {
	// Define the spec-mandated escape mappings
	replacer := strings.NewReplacer(
		`\,`, ",",
		`\;`, ";",
		`\\`, `\`,
		`\n`, "\n",
		`\N`, "\n",
	)
	// Replace the escape sequences in the input string with their corresponding characters
	// and return the unescaped text.
	return replacer.Replace(input)
}
