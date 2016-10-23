// Test file for typ65conv.go
// Scot W. Stevenson <scot.stevenson@gmail.com>
// First version 16. September 2016
// This version 20. September 2016

package main

import "testing"

// Test for empty and whitespace strings removed as we have already removed
// these cases in the main program
func TestFirstToUpper(t *testing.T) {
	var tests = []struct {
		input string
		want  string
	}{
		// {"", ""},
		// {" ", " "},
		{".end", ".END"},
		{"mr Robot", "MR Robot"},
		{"    .end", "    .END"},
		{"1234", "1234"},
		{"one fish two fish", "ONE fish two fish"},
	}

	for _, test := range tests {
		if got := firstToUpper(test.input); got != test.want {
			t.Errorf("firstToUpper(%q) = %v", test.input, got)
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

// TODO add more tests
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
