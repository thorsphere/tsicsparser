// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsicsparser

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/thorsphere/tserr"
)

// keyValue is a simple struct that holds a key-value pair.
type keyValue struct {
	Key   string // The key of the key-value pair.
	Value string // The value of the key-value pair.
}

// findCalDef searches for the "BEGIN:VCALENDAR" keyword in the input stream using the provided scanner.
func findCalDef(scanner *bufio.Scanner) error {
	// Search for the "BEGIN:VCALENDAR" keyword in the input stream.
	keyword := "BEGIN:VCALENDAR"
	// Scan through the input stream until we find the keyword or reach the end of the stream.
	for scanner.Scan() {
		// Read the current line from the scanner.
		line := scanner.Text()
		// Ignore empty lines.
		if len(line) == 0 {
			continue
		}
		// If the current line matches the keyword, we have found the calendar definition.
		if line == keyword {
			return nil
		}
	}
	// If we reach here, it means we did not find the keyword in the input stream.
	return tserr.NotFound(keyword)
}

func parseCalendar(scanner *ICSScanner) (*Calendar, error) {
	// Create a new Calendar struct to hold the parsed calendar information.
	var cal Calendar
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
		// Process the key-value pair (this is a simplified example, you might want
		// to do more with the actual values).
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
				// Call the parseEvent function to parse the event component and add it to the Calendar struct.
				event, err := parseEvent(scanner)
				// If there is an error while parsing the event, return the error.
				if err != nil {
					return nil, err
				}
				// Append the parsed event to the Events slice in the Calendar struct.
				cal.Events = append(cal.Events, event)
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
				return &cal, nil
			default:
				return nil, tserr.InvalidFormat("Unexpected END:" + parts.Value)
			}
		default:
			continue // Ignore other keys.
		}
	}
	return nil, tserr.InvalidFormat("Unexpected end of input while parsing calendar")
}

func parseEvent(scanner *ICSScanner) (Event, error) {
	return Event{}, nil
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
