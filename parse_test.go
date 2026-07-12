// Copyright (c) 2026 thorsphere.
// All Rights Reserved. Use is governed with GNU Affero General Public License v3.0
// that can be found in the LICENSE file.
package tsicsparser_test

import (
	"testing"

	"github.com/thorsphere/tserr"
	"github.com/thorsphere/tsicsparser" // Import the tsicsparser package to test the parseTimezone function.
)

// TestParseProdId tests the parseProdId function by providing a sample product identifier string
// and verifying that the parsed output matches the expected values. It ensures that the parser
// correctly extracts the registered flag, organization, product, and language components from
// the product identifier string.
func TestParseProdId(t *testing.T) {
	// Define the expected components of the product identifier.
	r := "+"
	o := "Thorsphere"
	p := "tsicsparser"
	l := "en"
	// Define a sample product identifier string to test the parseProdId function.
	pid := r + "//" + o + "//" + p + "//" + l
	// Call the parseProdId function to parse the product identifier string.
	prodId, err := tsicsparser.ParseProdId(pid)
	// If there is an error during parsing, report it and stop the test.
	if err != nil {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: pid, Err: err}))
	}
	// Check if the parsed product identifier is registered (the first component should be "+").
	if !prodId.Registered {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: pid, Err: tserr.InvalidFormat("Expected registered ProdID")}))
	}
	// Check if the organization field matches the expected value.
	if prodId.Organisation != o {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: prodId.Organisation, Err: tserr.InvalidFormat("Unexpected Organisation")}))
	}
	// Check if the product field matches the expected value.
	if prodId.Product != p {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: prodId.Product, Err: tserr.InvalidFormat("Unexpected Product")}))
	}
	// Check if the language field matches the expected value.
	if prodId.Language != l {
		t.Fatal(tserr.Op(&tserr.OpArgs{Op: "ParseProdId", Fn: prodId.Language, Err: tserr.InvalidFormat("Unexpected Language")}))
	}
}

// TestParseProdIdErr tests the parseProdId function with invalid product identifier strings
// to ensure that it correctly returns errors for malformed input. It verifies that the parser
// handles cases where the registered flag or product identifier format is incorrect.
func TestParseProdIdErr(t *testing.T) {
	// Define a set of test cases with invalid product identifier strings to test error handling.
	tests := []struct {
		name  string
		input string
	}{
		{name: "Invalid Registered Flag", input: "bad//Thorsphere//tsicsparser//en"},
		{name: "Invalid ProdID", input: "+//Thorsphere//tsicsparser"},
	}
	// Iterate over the test cases and run each one as a subtest.
	for _, tt := range tests {
		// Run each test case as a subtest to isolate failures and provide better reporting.
		t.Run(tt.name, func(t *testing.T) {
			// Call the parseProdId function to parse the product identifier string.
			_, err := tsicsparser.ParseProdId(tt.input)
			// If there is an error during parsing, report it and stop the test.
			if err == nil {
				t.Fatal(tserr.NilFailed("parseProdID"))
			}
		},
		)
	}
}
