// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed by the Functional Source License v1.1
// (FSL-1.1-ALv2) that can be found in the LICENSE file.
package tsicsparser

import (
	"fmt"
	"strings"

	"github.com/thorsphere/tserr"
)

// Calendar represents a calendar with events.
type Calendar struct {
	ProdId   ProdId   // Prodid is the product identifier for the calendar.
	Version  string   // Version is the version of the calendar format.
	Calscale string   // Calscale is the calendar scale used (e.g., "GREGORIAN").
	Method   string   // Method is the method used for the calendar (e.g., "PUBLISH").
	Summary  string   // Summary is a brief description of the calendar.
	Timezone Timezone // Timezone represents the timezone information for the calendar.
	Events   []Event  // Events is a slice of Event structs representing the events in the calendar.
}

// Prodid represents the product identifier for the calendar.
type ProdId struct {
	Registered   bool   // Registered indicates whether the product identifier is registered.
	Organisation string // Organization is the name of the organization associated with the product identifier.
	Product      string // Product is the name of the product associated with the product identifier.
	Language     string // Language is the ISO 639-1 language code associated with the product identifier.
}

// keyValue is a simple struct that holds a key-value pair.
type keyValue struct {
	Key   string // The key of the key-value pair.
	Value string // The value of the key-value pair.
}

// parseCalendar parses a VCALENDAR component from the scanner.
//
// Parsing is two-pass to remove an ordering dependency that RFC 5545 does
// not impose: a VEVENT may reference a TZID whose VTIMEZONE appears later
// in the stream. Because the ICSScanner is forward-only, VEVENT blocks
// are buffered as raw lines during the first pass (alongside immediate
// parsing of VTIMEZONE blocks) and only handed to parseEvent in the
// second pass, once every VTIMEZONE has been collected into cal.Timezone.
func parseCalendar(scanner *ICSScanner) (*Calendar, error) {
	// Create a new Calendar struct to hold the parsed calendar information.
	var cal Calendar
	// Initialize a flag to indicate whether we have started parsing the calendar.
	calStarted := false
	// rawEvents holds the raw line slices of each VEVENT block seen
	// during pass 1 (the lines between BEGIN:VEVENT and END:VEVENT,
	// exclusive). They are parsed in pass 2 via parseBufferedEvents.
	var rawEvents [][]string
	// Scan through the input stream to read the calendar header information.
	for scanner.Scan() {
		// Read the current line from the scanner.
		line := scanner.Text()
		// Ignore empty lines.
		if len(line) == 0 {
			continue
		}
		// Split the line into key-value pairs based on the first colon.
		parts, err := splitKeyValue(line)
		// If there is an error while splitting the line, return the error.
		if err != nil {
			return nil, err
		}
		// If we have not yet started parsing the calendar and
		// the current line indicates the beginning of a calendar,
		// set the calStarted flag to true and continue to the next line.
		if !calStarted {
			// Check if the current line indicates the beginning of a calendar.
			if parts.Key == "BEGIN" && parts.Value == "VCALENDAR" {
				// Set the calStarted flag to true to indicate that we have started parsing the calendar.
				calStarted = true
			}
			// Continue to the next line in the input stream.
			continue
		}
		// If we have started parsing the calendar, we need to handle the different keys
		// in the calendar header.
		switch parts.Key {
		case "PRODID": // If the key is "PRODID", we need to parse the product identifier information.
			// Call the parseProdID function to parse the PRODID field and set the corresponding
			// fields in the Calendar struct.
			prodID, err := parseProdID(parts.Value)
			// If there is an error while parsing the PRODID field, return the error.
			if err != nil {
				return nil, err
			}
			// Set the ProdId field in the Calendar struct to the parsed product identifier.
			cal.ProdId = prodID
		case "VERSION": // If the key is "VERSION", set the Version field in the Calendar struct.
			cal.Version = parts.Value
		case "CALSCALE": // If the key is "CALSCALE", set the Calscale field in the Calendar struct.
			cal.Calscale = parts.Value
		case "METHOD": // If the key is "METHOD", set the Method field in the Calendar struct.
			cal.Method = parts.Value
		case "SUMMARY": // If the key is "SUMMARY", set the Summary field in the Calendar struct.
			cal.Summary = parts.Value
		case "BEGIN": // If the key is "BEGIN", we need to handle the beginning of a new component.
			switch parts.Value {
			case "VEVENT": // If the value is "VEVENT", we are starting a new event component.
				// Pass 1: do not parse yet — buffer the raw lines so the
				// event can be resolved against cal.Timezone in pass 2
				// regardless of where its VTIMEZONE appears in the stream.
				raw, err := collectRawBlock(scanner, "VEVENT")
				// If there is an error while parsing the event, return the error.
				if err != nil {
					return nil, err
				}
				// Append the raw lines of the event component to the rawEvents slice for later parsing.
				rawEvents = append(rawEvents, raw)
			case "VTIMEZONE": // If the value is "VTIMEZONE", we are starting a new timezone component.
				// Call the parseTimezone function to parse the timezone component and add it to the Calendar struct.
				timezone, err := parseTimezone(scanner)
				// If there is an error while parsing the timezone, return the error.
				if err != nil {
					return nil, err
				}
				// Set the Timezone field in the Calendar struct to the parsed timezone.
				cal.Timezone = timezone
			default:
				continue // Ignore other components.
			}
		case "END": // If the key is "END", we need to handle the end of a component.
			switch parts.Value {
			case "VEVENT": // If the value is "VEVENT", we have reached the end of an event component.
				return nil, tserr.InvalidFormat("Unexpected END:VEVENT without matching BEGIN:VEVENT")
			case "VTIMEZONE": // If the value is "VTIMEZONE", we have reached the end of a timezone component.
				return nil, tserr.InvalidFormat("Unexpected END:VTIMEZONE without matching BEGIN:VTIMEZONE")
			case "VCALENDAR": // If the value is "VCALENDAR", we have reached the end of the calendar component.
				// Pass 2: all VTIMEZONE blocks have been collected. Parse the buffered VEVENT blocks now.
				if err := parseBufferedEvents(&cal, rawEvents); err != nil {
					// If there is an error while parsing the buffered events, return the error.
					return nil, err
				}
				// Return the parsed Calendar struct and nil error to indicate successful parsing.
				return &cal, nil
			default:
				// If we encounter an unexpected END key, return an error indicating invalid format.
				return nil, tserr.InvalidFormat("Unexpected END:" + parts.Value)
			}
		default:
			continue // Ignore other keys.
		}
	}
	// If we reach here and calStarted is false, it means we have reached the end of the input stream
	// without finding the "BEGIN:VCALENDAR" keyword.
	if !calStarted {
		return nil, tserr.NotFound("BEGIN:VCALENDAR")
	}
	// EOF without END:VCALENDAR. Parse buffered events first so any
	// event-level error surfaces before the missing-close error.
	if err := parseBufferedEvents(&cal, rawEvents); err != nil {
		return nil, err
	}
	// If we reach here, it means we have reached the end of the input stream
	// without finding the "END:VCALENDAR" keyword.
	return nil, tserr.InvalidFormat("Unexpected end of input while parsing calendar")
}

// collectRawBlock reads lines from scanner until the matching END:<block>
// line, returning the raw lines in between (the caller has already
// consumed BEGIN:<block>). It mirrors the line-handling skeleton of
// parseEvent — skipping empty lines, propagating splitKeyValue errors,
// and rejecting any END:<other> as "unexpected END inside <block>" — so
// that structural errors are detected during pass 1 at the same point
// parseEvent would have detected them, while leaving DTSTART/DTEND/
// DURATION resolution to pass 2. Returns tserr.NotFound("END:<block>")
// if the stream ends before the close, matching parseEvent.
func collectRawBlock(scanner *ICSScanner, block string) ([]string, error) {
	// Initialize a slice to hold the raw lines of the block.
	var lines []string
	// Scan through the input stream to read the lines of the block until we find the matching END:<block> line.
	for scanner.Scan() {
		// Read the current line from the scanner.
		line := scanner.Text()
		// Ignore empty lines.
		if len(line) == 0 {
			continue
		}
		// Split the line into key-value pairs based on the first colon.
		parts, err := splitKeyValue(line)
		// If there is an error while splitting the line, return the error.
		if err != nil {
			return nil, err
		}
		// If we encounter the matching END:<block> line, return the collected lines and nil error.
		if parts.Key == "END" {
			// If the END key matches the expected block, return the collected lines and nil error.
			if parts.Value == block {
				return lines, nil
			}
			// If we encounter an unexpected END key, return an error indicating invalid format.
			return nil, tserr.InvalidFormat(fmt.Sprintf("unexpected END:%s inside %s", parts.Value, block))
		}
		// Append the current line to the lines slice to collect the raw lines of the block.
		lines = append(lines, line)
	}
	// If we reach here, it means we have reached the end of the input stream
	if err := scanner.Err(); err != nil {
		// If there is an error while scanning the input stream, return the error.
		return nil, err
	}
	// If we reach here, it means we have reached the end of the input stream
	// without finding the matching "END:<block>" keyword.
	return nil, tserr.NotFound("END:" + block)
}

// parseBufferedEvents is pass 2 of parseCalendar: it re-scans each raw
// VEVENT block collected during pass 1 and hands it to parseEvent with
// the fully-resolved calendar timezone. The raw lines are reconstructed
// with a trailing END:VEVENT so parseEvent sees the same component close
// it would have seen inline. The ICSScanner is forward-only, so each
// block gets its own scanner over the reconstructed string; line folding
// is a no-op on the second scanner because pass 1 already unfolded the
// lines.
func parseBufferedEvents(cal *Calendar, rawEvents [][]string) error {
	// Iterate over each raw event block collected during pass 1.
	for _, raw := range rawEvents {
		// Reconstruct the raw event block by joining the lines with newline characters and appending "END:VEVENT" to indicate the end of the event component.
		body := strings.Join(raw, "\n") + "\nEND:VEVENT"
		// Create a new ICSScanner for the raw event block, specifying "VEVENT" as the component type.
		s := NewICSScanner(strings.NewReader(body), "VEVENT")
		// Parse the event using the parseEvent function, passing in the scanner and the calendar's timezone.
		event, err := parseEvent(s, cal.Timezone)
		// If there is an error while parsing the event, return the error.
		if err != nil {
			return err
		}
		// Append the parsed event to the Events slice in the Calendar struct.
		cal.Events = append(cal.Events, event)
	}
	// If we reach here, it means all buffered events have been successfully parsed and added to the Calendar struct.
	return nil
}

func parseProdID(value string) (ProdId, error) {
	// Split the value of the PRODID field into its components using "//" as the delimiter.
	parts := strings.SplitN(value, "//", 4)
	// If the number of components is less than 4, return an error indicating invalid format.
	if len(parts) < 4 {
		return ProdId{}, tserr.InvalidFormat(value)
	}
	// Check if the first component is either "+" or "-", indicating whether the product identifier is registered.
	if parts[0] != "+" && parts[0] != "-" {
		return ProdId{}, tserr.InvalidFormat(fmt.Sprintf("Unexpected first ProdID component %s", parts[0]))
	}
	// Create a new ProdId struct and populate its fields based on the parsed components. The first
	// component indicates whether the product identifier is registered (if it starts with "+"),
	// the second component is the organization, the third component is the product,
	// and the fourth component is the language.
	return ProdId{
		Registered:   parts[0] == "+",
		Organisation: parts[1],
		Product:      parts[2],
		Language:     parts[3],
	}, nil
}

// splitKeyValue splits a line into a key-value pair based on the first colon.
func splitKeyValue(line string) (*keyValue, error) {
	// Split the line into key-value pairs based on the first colon.
	parts := strings.SplitN(line, ":", 2)
	// If the line does not contain a colon, return an error indicating invalid format.
	if len(parts) != 2 {
		return nil, tserr.InvalidFormat(line)
	}
	// Extract the key and value from the split parts.
	key := parts[0]
	// The value is the part after the colon.
	value := parts[1]
	// Return a new keyValue struct containing the extracted key and value.
	return &keyValue{Key: key, Value: value}, nil
}
