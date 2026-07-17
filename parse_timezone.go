// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsicsparser

// This file contains functions to parse VTIMEZONE components from ICS files.
import (
	// For reading lines from the ICS input stream
	"fmt"     // For formatting error messages
	"strconv" // For converting string representations of numbers to integers
	"strings" // For string manipulation functions like Split and SplitN
	"time"    // For working with time and date values

	"github.com/thorsphere/tserr"   // For custom error handling and reporting
	"github.com/thorsphere/tstable" // For creating and formatting tables for output
)

// RRule represents a recurrence rule for timezone transitions.
type RRule struct {
	Freq    string // e.g., "YEARLY"
	ByMonth int    // 1-12
	ByDay   string // e.g., "1SU", "-1SU"
}

// RuleType distinguishes between standard and daylight saving rules.
type RuleType int

// Constants for RuleType
const (
	Standard RuleType = iota
	Daylight
)

// TimezoneRule represents a single timezone transition rule (DAYLIGHT or STANDARD).
type TimezoneRule struct {
	Type         RuleType  // Type of the rule (Standard or Daylight)
	DTStart      time.Time // Local time of the first effective date of the rule
	TZOffsetFrom int       // Offset in seconds BEFORE the rule takes effect
	TZOffsetTo   int       // Offset in seconds AFTER the rule takes effect
	TZName       string    // Abbreviation, e.g., "CEST" or "CET"
	RRule        *RRule    // Optional, missing for fixed transitions (only DTSTART)
}

// Timezone represents a timezone with its identifier and associated rules.
type Timezone struct {
	TZID  string         // Timezone identifier, e.g., "US-Eastern"
	Rules []TimezoneRule // List of rules for the timezone, including standard and daylight saving time rules
}

// parseTimezone parses a VTIMEZONE component from the scanner, consuming lines
// until END:VTIMEZONE is reached. It handles both DAYLIGHT and STANDARD
// sub-components including DTSTART, TZOFFSETFROM, TZOFFSETTO, TZNAME, and RRULE.
func parseTimezone(scanner *ICSScanner) (Timezone, error) {
	// Initialize a Timezone struct to hold the parsed timezone information.
	var tz Timezone
	// Initialize a pointer to the current TimezoneRule being parsed (either DAYLIGHT or STANDARD).
	var currentRule *TimezoneRule
	// Loop through the lines of the scanner until END:VTIMEZONE is encountered.
	for scanner.Scan() {
		// Read the next line from the scanner.
		line := scanner.Text()
		// Skip empty lines to avoid unnecessary processing.
		if len(line) == 0 {
			continue
		}
		// Split the line into key and value parts using the splitKeyValue function.
		parts, err := splitKeyValue(line)
		// If there is an error splitting the line, return an error indicating the issue.
		if err != nil {
			return Timezone{}, err
		}
		// Use a switch statement to handle different keys in the VTIMEZONE component.
		switch parts.Key {
		// Handle the TZID property, which specifies the timezone identifier.
		case "TZID":
			tz.TZID = parts.Value
			// Handle the BEGIN of a timezone rule (DAYLIGHT or STANDARD).
		case "BEGIN":
			// If the BEGIN is for a DAYLIGHT or STANDARD rule, initialize a new TimezoneRule.
			switch parts.Value {
			// If the BEGIN is for a DAYLIGHT rule,
			// create a new TimezoneRule with Type set to Daylight.
			case "DAYLIGHT":
				currentRule = &TimezoneRule{Type: Daylight}
			// If the BEGIN is for a STANDARD rule,
			// create a new TimezoneRule with Type set to Standard.
			case "STANDARD":
				currentRule = &TimezoneRule{Type: Standard}
			// If the BEGIN is for any other value, return an error indicating unexpected BEGIN.
			default:
				return Timezone{}, tserr.InvalidFormat(fmt.Sprintf("unexpected BEGIN:%s inside VTIMEZONE", parts.Value))
			}
		// Handle the END of a timezone rule or the entire VTIMEZONE component.
		case "END":
			// If the END is for a DAYLIGHT or STANDARD rule,
			// append the current rule to the timezone's rules.
			switch parts.Value {
			// If the END is for a DAYLIGHT or STANDARD rule,
			// append the current rule to the timezone's rules.
			case "DAYLIGHT":
				// If the current rule is not nil and is of type Daylight,
				// append it to the timezone's rules and reset currentRule.
				if currentRule != nil && currentRule.Type == Daylight {
					// Append the current rule to the timezone's rules.
					tz.Rules = append(tz.Rules, *currentRule)
					// Reset currentRule to nil to indicate that we are no longer processing a rule.
					currentRule = nil
				} else {
					// If the current rule is nil or not of type Daylight, return an error indicating that there is no matching BEGIN:DAYLIGHT.
					return Timezone{}, tserr.InvalidFormat("END:DAYLIGHT without matching BEGIN:DAYLIGHT")
				}
			// If the END is for a STANDARD rule, perform similar checks and actions as for DAYLIGHT.
			case "STANDARD":
				// If the current rule is not nil and is of type Standard,
				// append it to the timezone's rules and reset currentRule.
				if currentRule != nil && currentRule.Type == Standard {
					// Append the current rule to the timezone's rules.
					tz.Rules = append(tz.Rules, *currentRule)
					// Reset currentRule to nil to indicate that we are no longer processing a rule.
					currentRule = nil
				} else {
					// If the current rule is nil or not of type Standard, return an error indicating that there is no matching BEGIN:STANDARD.
					return Timezone{}, tserr.InvalidFormat("END:STANDARD without matching BEGIN:STANDARD")
				}
			// If the END is for the entire VTIMEZONE component, return the parsed timezone.
			case "VTIMEZONE":
				// Return the parsed timezone and a nil error,
				// indicating successful parsing of the VTIMEZONE component.
				return tz, nil
			// If the END is for any other value, return an error indicating unexpected END.
			default:
				return Timezone{}, tserr.InvalidFormat(
					fmt.Sprintf("unexpected END:%s inside VTIMEZONE", parts.Value))
			}
		// Handle DTSTART for timezone transitions.
		case "DTSTART":
			// If the current rule is not nil, parse the DTSTART value and
			// assign it to the current rule.
			if currentRule != nil {
				// Parse the DTSTART value and convert it to time.Time. The isUTC flag is not used
				// because the DTSTART field is always in local time.
				t, _, err := parseICSDateTime(parts.Value)
				// If there is an error parsing the DTSTART value,
				// return an error indicating the issue.
				if err != nil {
					return Timezone{}, err
				}
				// Assign the parsed time value to the current rule's DTStart field.
				currentRule.DTStart = t
			} else {
				// If the current rule is nil, return an error indicating that the DTSTART is not associated with any rule.
				return Timezone{}, tserr.InvalidFormat("DTSTART without a current rule (missing BEGIN:DAYLIGHT or BEGIN:STANDARD)")
			}
		// Handle TZOFFSETFROM for timezone transitions.
		case "TZOFFSETFROM":
			// If the current rule is not nil, parse the TZOFFSETFROM value and
			// assign it to the current rule.
			if currentRule != nil {
				// Parse the TZOFFSETFROM value and convert it to seconds.
				offset, err := parseOffset(parts.Value)
				// If there is an error parsing the TZOFFSETFROM value,
				// return an error indicating the issue.
				if err != nil {
					return Timezone{}, err
				}
				// Assign the parsed offset value to the current rule's TZOffsetFrom field.
				currentRule.TZOffsetFrom = offset
			} else {
				// If the current rule is nil, return an error indicating that the TZOFFSETFROM is not associated with any rule.
				return Timezone{}, tserr.InvalidFormat("TZOFFSETFROM without a current rule (missing BEGIN:DAYLIGHT or BEGIN:STANDARD)")
			}
		// Handle TZOFFSETTO for timezone transitions.
		case "TZOFFSETTO":
			// If the current rule is not nil, parse the TZOFFSETTO value and
			// assign it to the current rule.
			if currentRule != nil {
				// Parse the TZOFFSETTO value and convert it to seconds.
				offset, err := parseOffset(parts.Value)
				// If there is an error parsing the TZOFFSETTO value,
				// return an error indicating the issue.
				if err != nil {
					return Timezone{}, err
				}
				// Assign the parsed offset value to the current rule's TZOffsetTo field.
				currentRule.TZOffsetTo = offset
			} else {
				// If the current rule is nil, return an error indicating that the TZOFFSETTO is not associated with any rule.
				return Timezone{}, tserr.InvalidFormat("TZOFFSETTO without a current rule (missing BEGIN:DAYLIGHT or BEGIN:STANDARD)")
			}
		// Handle TZNAME for timezone abbreviations in timezone transitions.
		case "TZNAME":
			// If the current rule is not nil, assign the TZNAME value to the current rule.
			if currentRule != nil {
				currentRule.TZName = parts.Value
			} else {
				// If the current rule is nil, return an error indicating that the TZNAME is not associated with any rule.
				return Timezone{}, tserr.InvalidFormat("TZNAME without a current rule (missing BEGIN:DAYLIGHT or BEGIN:STANDARD)")
			}
		// Handle RRULE for recurrence rules in timezone transitions.
		case "RRULE":
			// If the current rule is not nil, parse the RRULE value and assign it to the current rule.
			if currentRule != nil {
				// Parse the RRULE value and assign it to the current rule.
				rrule, err := parseRRule(parts.Value)
				// If there is an error parsing the RRULE, return an error indicating the issue.
				if err != nil {
					return Timezone{}, err
				}
				// Assign the parsed RRULE to the current rule's RRule field.
				currentRule.RRule = rrule
			} else {
				// If the current rule is nil, return an error indicating that the RRULE is not associated with any rule.
				return Timezone{}, tserr.InvalidFormat("RRULE without a current rule (missing BEGIN:DAYLIGHT or BEGIN:STANDARD)")
			}
		}
	}
	// Check for any scanning errors that may have occurred during the parsing process.
	if err := scanner.Err(); err != nil {
		return Timezone{}, err
	}
	// END:VTIMEZONE was never encountered.
	// Return an error indicating that the VTIMEZONE component was not properly closed.
	return Timezone{}, tserr.NotFound("END:VTIMEZONE")
}

// parseOffset parses an ICS UTC offset string in ±HHMM format into seconds.
// Examples: "-0500" → -18000, "+0530" → 19800, "+0000" → 0.
func parseOffset(s string) (int, error) {
	// Return an error if the offset string does not have exactly 5 characters (±HHMM).
	if len(s) != 5 {
		return 0, tserr.InvalidFormat(fmt.Sprintf("invalid offset length: %s", s))
	}
	// Initialize the sign variable to 1 (positive) by default.
	sign := 1
	// Determine the sign of the offset based on the first character of the string.
	switch s[0] {
	case '-':
		sign = -1 // If the first character is '-', set sign to -1 for negative offset.
	case '+':
		// If the first character is '+', keep sign as 1 for positive offset.
	default:
		// If the first character is neither '-' nor '+', return an error indicating invalid format.
		return 0, tserr.InvalidFormat(fmt.Sprintf("invalid offset sign: %s", s))
	}
	// Parse the hours part of the offset string (characters 1 and 2) and convert it to an integer.
	hours, err := strconv.Atoi(s[1:3])
	// If there is an error parsing the hours part, return an error indicating invalid format.
	if err != nil {
		return 0, tserr.InvalidFormat(fmt.Sprintf("invalid offset hours: %s", s))
	}
	// Parse the minutes part of the offset string (characters 3 and 4) and convert it to an integer.
	minutes, err := strconv.Atoi(s[3:5])
	// If there is an error parsing the minutes part, return an error indicating invalid format.
	if err != nil {
		return 0, tserr.InvalidFormat(fmt.Sprintf("invalid offset minutes: %s", s))
	}
	// Calculate the total offset in seconds by converting hours and minutes to seconds,
	// applying the sign, and returning the result.
	return sign * (hours*3600 + minutes*60), nil
}

// parseRRule parses an RRULE value like "FREQ=YEARLY;BYMONTH=3;BYDAY=2SU".
func parseRRule(s string) (*RRule, error) {
	// Return an error if the RRULE string is empty, as it cannot be parsed.
	if s == "" {
		return nil, tserr.InvalidFormat("empty RRULE")
	}
	// Initialize an empty RRule struct to hold the parsed values.
	var rrule RRule
	// Split the RRULE string by semicolons to separate each key-value pair.
	// Iterate over each part of the RRULE string.
	for part := range strings.SplitSeq(s, ";") {
		// Split each part by the first "=" to separate the key and value.
		kv := strings.SplitN(part, "=", 2)
		// If the split does not result in exactly two parts (key and value),
		// return an error indicating invalid format.
		if len(kv) != 2 {
			return nil, tserr.InvalidFormat(fmt.Sprintf("invalid RRULE part: %s", part))
		}
		// Use a switch statement to handle different keys in the RRULE.
		switch kv[0] {
		// If the key is "FREQ", assign the value to the Freq field of the RRule struct.
		case "FREQ":
			rrule.Freq = kv[1]
		// If the key is "BYMONTH", convert the value to an integer and
		// assign it to the ByMonth field.
		case "BYMONTH":
			// Convert the value to an integer representing the month (1-12).
			month, err := strconv.Atoi(kv[1])
			// If there is an error converting the value to an integer,
			// return an error indicating invalid format.
			if err != nil {
				return nil, tserr.InvalidFormat(fmt.Sprintf("invalid RRULE part: %s", part))
			}
			// Assign the parsed month value to the ByMonth field of the RRule struct.
			rrule.ByMonth = month
		// If the key is "BYDAY", assign the value to the ByDay field of the RRule struct.
		case "BYDAY":
			rrule.ByDay = kv[1]
		}
	}
	// Return the parsed RRule struct and a nil error, indicating successful parsing.
	return &rrule, nil
}

// parseICSDateTime parses an ICS datetime string in the form "20070311T020000" or "20070311T020000Z"
// into a time.Time (interpreted as a local time without timezone info) and
// a boolean flag indicating if the datetime is already UTC.
func parseICSDateTime(s string) (time.Time, bool, error) {
	// Initialize a layout string to hold the datetime layout
	var layout string
	// Check if the datetime is already UTC by checking if it ends with "Z"
	isUTC := strings.HasSuffix(s, "Z")
	// If the datetime is already UTC, use a different layout
	if isUTC {
		layout = "20060102T150405Z"
	} else { // Otherwise, use the default layout
		layout = "20060102T150405"
	}
	// Use the time.Parse function to parse the ICS datetime string using the specified layout.
	t, err := time.Parse(layout, s)
	// If there is an error during parsing, return a zero time.Time value, the isUTC flag, and
	// an error indicating invalid format.
	if err != nil {
		return time.Time{}, false, tserr.InvalidFormat(fmt.Sprintf("invalid datetime format: %s", s))
	}
	// Return the parsed time.Time value, the isUTC flag, and a nil error, indicating successful parsing.
	return t, isUTC, nil
}

// String returns a string representation of the RuleType (either "STANDARD" or "DAYLIGHT").
func (rt RuleType) String() string {
	// Use a switch statement to return the string representation of the RuleType.
	switch rt {
	// If the RuleType is Standard, return "STANDARD".
	case Standard:
		return "STANDARD"
	// If the RuleType is Daylight, return "DAYLIGHT".
	case Daylight:
		return "DAYLIGHT"
	// If the RuleType is unknown, return "<unknown>".
	default:
		return "<unknown>"
	}
}

// String returns a formatted string representation of the Timezone struct,
// including its TZID and associated rules. It uses the tstable package to create a table-like output.
func (tz Timezone) String() string {
	// Create a new table
	tbl, err := tstable.New([]string{"Timezone ID", tz.TZID})
	// If there is an error, return an empty string
	if err != nil {
		return fmt.Sprintf("<error creating table: %v>", err)
	}
	// If there are no rules, indicate that in the table and return the string representation.
	if len(tz.Rules) == 0 {
		tbl.AddRow([]string{"Rules", "<none>"})
		return tbl.String()
	}
	// Add rows for each rule in the timezone
	for i, rule := range tz.Rules {
		// Add a separator between rules for better readability
		if i > 0 {
			tbl.AddSeparator()
		}
		// Add rows for the rule's details
		tbl.AddRow([]string{"Rule Type", rule.Type.String()})
		tbl.AddRow([]string{"DTStart", rule.DTStart.Format(time.RFC3339)})
		tbl.AddRow([]string{"TZOffsetFrom", fmt.Sprintf("%+d", rule.TZOffsetFrom)})
		tbl.AddRow([]string{"TZOffsetTo", fmt.Sprintf("%+d", rule.TZOffsetTo)})
		tbl.AddRow([]string{"TZName", rule.TZName})
		// If the rule has an RRule, add its details to the table
		if rule.RRule != nil {
			tbl.AddRow([]string{"RRule Freq", rule.RRule.Freq})
			tbl.AddRow([]string{"RRule ByMonth", fmt.Sprintf("%d", rule.RRule.ByMonth)})
			tbl.AddRow([]string{"RRule ByDay", rule.RRule.ByDay})
		}
	}
	// Return the formatted table as a string
	return tbl.String()
}
