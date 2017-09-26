// Test file for typ65conv.go
// Scot W. Stevenson <scot.stevenson@gmail.com>
// First version 16. Sep 2016
// This version 26. Sep 2017

package main

import "testing"

func TestConvertNumber(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		// {"", ""},
		// {" ", " "},
		{"00:0000", "$000000"},
		{"00.0000", "$000000"},
		{"%11110000", "%11110000"},
		{"$00", "$00"},
		{"&10", "10"},
		{"&00", "00"},
		{"0xa0", "$a0"},
		{"0x00", "$00"},
		{"1000", "$1000"},
		{"0000", "$0000"},
		{"tali", "tali"},
		{"dead", "$dead"},   // should be converted to number
		{"0dead", "$0dead"}, // should be a number
		{"tali", "tali"},
	}

	for _, test := range tests {
		if got := convertNumber(test.input); got != test.want {
			t.Errorf("convertNumber(%q) = %v", test.input, got)
		}
	}
}

// Test for empty lines and comments not necessary because are assumed to have
// been delt with
func TestHasLabel(t *testing.T) {
	var tests = []struct {
		input string
		want  bool
	}{
		// {"", ""},
		// {" ", " "},
		{"label", true},
		{"   label", false},
		{"    .end", false},
		{"label nop", true},
	}

	for _, test := range tests {
		if got := hasLabel(test.input); got != test.want {
			t.Errorf("hasLabel(%q) = %v", test.input, got)
		}
	}
}

func TestIsComment(t *testing.T) {
	var tests = []struct {
		input string
		want  bool
	}{
		{"", false},           // empty line
		{"tali", false},       // label
		{" .mpu 6502", false}, // directive
		{" tax", false},       // opcode
		{";", true},           // comment at beginning of line
		{" ;", true},          // comment after indent
	}

	for _, test := range tests {
		if got := isComment(test.input); got != test.want {
			t.Errorf("isComment(%q) = %v", test.input, got)
		}
	}
}

func TestIsDirective(t *testing.T) {
	var tests = []struct {
		input string
		want  bool
	}{
		{"tali", false},      // label
		{";", false},         // comment
		{" .mpu 6502", true}, // directive
		{" tax", false},      // opcode
	}

	for _, test := range tests {
		if got := isDirective(test.input); got != test.want {
			t.Errorf("isDirective(%q) = %v", test.input, got)
		}
	}
}

// Test for isOpcode, currently does not work because requires call to main
// routine to load opcodes from JSON file
/*
func TestIsOpcode(t *testing.T) {
	var tests = []struct {
		input string
		want  bool
	}{
		{"", false},
		{"superfrog", false},
		{"tax", true},
		{"adc.l", true},
		{".data", false},
	}

	for _, test := range tests {
		if got := isOpcode(test.input); got != test.want {
			t.Errorf("isOpcode(%q) = %v", test.input, got)
		}
	}
}
*/

func TestIsEmpty(t *testing.T) {
	var tests = []struct {
		input string
		want  bool
	}{
		{"tali", false},
		{"", true},
		{" ", true},
		{"\t", true},
	}

	for _, test := range tests {
		if got := isEmpty(test.input); got != test.want {
			t.Errorf("isEmpty(%q) = %v", test.input, got)
		}
	}
}

func TestMergeLabel(t *testing.T) {
	type ip struct {
		line  string
		label string
	}
	var tests = []struct {
		input ip
		want  string
	}{
		{ip{"label ; stuff", "xxxxx"}, "xxxxx ; stuff"},
		{ip{"label", "label:"}, "label:"},
	}

	for _, test := range tests {
		if got := mergeLabel(test.input.line, test.input.label); got != test.want {
			t.Errorf("mergeLabel(%q) = %v", test.input, got)
		}
	}
}

func TestRemoveSeparators(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		{"0000", "0000"},
		{"00:00", "0000"},
		{"00.00", "0000"},
		{"0:0.00", "0000"},
	}

	for _, test := range tests {
		if got := removeSeparators(test.input); got != test.want {
			t.Errorf("removeSeparators(%q) = %v", test.input, got)
		}
	}
}
