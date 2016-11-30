// typ65conv - Convert Typist's Assembler Code to Traditional Formats
// Scot W. Stevenson <scot.stevenson@gmail.com>
// First version: 16. Sep 2016
// This version: 30. Nov 2016

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// We keep line numbers associated with each line while we work on things for
// the error codes and to be able to store them as the last step
type workline struct {
	linenumber int
	payload    string
}

const (
	OPCODELIST = "opcodes.json"
)

// Opcode conversion data is stored in a JSON file
type OpcData struct {
	OldMnem string `json: "oldmnem"`
	Size    int    `json: "size"`
}

var Opcodes struct {
	Table map[string]OpcData `json: "table"`
}

// We limit the number of concurrent processes to the number of logical cores
// because we don't want to totally bog down the machine
var (
	maxworkers = runtime.GOMAXPROCS(0)
	raw        []string
	processed  []workline

	upperOpcs  = flag.Bool("ou", false, "Convert opcodes to upper case")
	labelColon = flag.Bool("lc", false, "Add colon to labels")
	input      = flag.String("i", "", "Input file (REQUIRED)")
	output     = flag.String("o", "typ65conv.asm", "Output file (default 'typ65conv.asm')")

	opcodeindent = strings.Repeat(" ", 16)
)

// Defintions for sorting, see https://golang.org/pkg/sort/
// We sort by line numbers
type byLine []workline

func (a byLine) Len() int           { return len(a) }
func (a byLine) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLine) Less(i, j int) bool { return a[i].linenumber < a[j].linenumber }

// convertNumber takes a string that is the opcode of an typist assembler
// instruction and returns a string converted for its traditional counterpart.
func convertNumber(s string) string {

	s = strings.TrimSpace(s)
	s = removeSeparators(s)

	if strings.HasPrefix(s, "%") {
		return s
	}

	if strings.HasPrefix(s, "&") {
		return strings.TrimPrefix(s, "&")
	}

	if strings.HasPrefix(s, "$") {
		return s
	}

	// Use TrimPrefix instead of TrimLeft or "0x00" will fail
	if strings.HasPrefix(s, "0x") {
		return "$" + strings.TrimPrefix(s, "0x")
	}

	// Check if string is a valid hex number. We have to use 32 as size of
	// numbers in bits because Go doesn't support 24 bits. Note this means
	// that DEAD etc will be considered a number, not a symbol
	if _, err := strconv.ParseInt(s, 16, 32); err == nil {
		return "$" + s
	}

	// Assume symbol or label
	return s
}

// firstToUpper takes a string and converts the first word to uppercase,
// returning the rest of the string otherwise unchanged. Splits at the first
// space character. Used to convert both directives and opcodes to upper case.
// Assumes that s is neither empty nor whitespace only
func firstToUpper(s string) string {
	wt := strings.TrimSpace(s) // otherwise whitespace will confuse strings.SplitN
	ws := strings.SplitN(wt, " ", 2)
	return strings.Replace(s, ws[0], strings.ToUpper(ws[0]), 1)
}

// hasLabel takes a string. If the string does not start with whitespace it is
// assumed to be a label and the bool true is returned, otherwise false. The
// function works regardless if there is anything else on the line since we only
// examine the first character. Assumes that comments and empty lines have been
// already checked for.
func hasLabel(s string) bool {
	return !strings.HasPrefix(s, " ")
}

// isComment takes a string. If the string starts with the comment character
// ";" after any whitespace, it returns the bool true, otherwise false
func isComment(s string) bool {
	cs := strings.TrimSpace(s)
	return strings.HasPrefix(cs, ";")
}

// isDirective takes a string. If the string starts with a dot as the first
// non-whitespace character, it is considered to be a directive. A bool is
// returned. Assumes that no label may start with a dot
func isDirective(s string) bool {
	cs := strings.TrimSpace(s)
	return strings.HasPrefix(cs, ".")
}

// isEmpty takes a string. If the string is empty or contains nothing but
// whitespace, return true, else return false
func isEmpty(s string) bool {
	cs := strings.TrimSpace(s)
	return cs == ""
}

// isOpcode takes a string. If the first word of the string is in the list of
// Typist Assembler Notation opcodes, it returns the bool true, else false.
func isOpcode(s string) bool {
	_, ok := Opcodes.Table[s]
	return ok
}

// mergeLabel takes two strings, the first a complete line with the original
// label, the second the new label. It returns a string of the same length as
// the original where the new label replaces the old one
// TODO write test routine
func mergeLabel(line, newlabel string) string {
	linelen := len(line)
	newlabellen := len(newlabel)

	// Handle single label in line with added colons separately
	if newlabellen >= linelen {
		return newlabel
	}

	return newlabel + line[newlabellen:]
}

// removeSeparators takes a string representation of a hex number and returns a
// string version with all legal separators removed. If there are no separators
// in the string, the original string is returned
func removeSeparators(s string) string {
	r := strings.NewReplacer(":", "", ".", "")
	return r.Replace(s)
}

// procLine takes a work line struct consisting of the line number and the
// payload string that is one of the source codes lines. It returns a new work
// line with the line number untouched and the payload string converted as
// instructed by switches. This function does the main work and will be later
// called as a goroutine
func procLine(jobs <-chan workline, results chan<- workline) {

	for j := range jobs {

		if isComment(j.payload) {
			results <- j
			continue
		}

		if isEmpty(j.payload) {
			results <- j
			continue
		}

		label := ""

		// TODO figure out a cleaner way of handling the colon addition,
		// this currently happens in three different places
		if hasLabel(j.payload) {

			// Assumes we have already checked for empty lines
			labelline := strings.Fields(j.payload)
			label = labelline[0]

			// If there is only the label in the line, we can deal
			// with it immediately
			if len(labelline) == 1 {

				if *labelColon {
					label = label + ":"
				}

				results <- workline{j.linenumber, label}
				continue
			}

			// If there is only the label and a comment, we can
			// reassemble the line right now and continue
			// TODO make sure there is enough whitespace left in the
			// cases where the label gets a colon added
			if isComment(labelline[1]) {

				if *labelColon {
					label = label + ":"
				}

				results <- workline{j.linenumber, mergeLabel(j.payload, label)}
				continue
			}

			// There is something else on the line we have to
			// process first, then we can come back and reassemble
			// stuff. We use the fact that label is not empty as a
			// flag
			j.payload = strings.Replace(j.payload, label, "", -1)

		}

		// Main switch statement returns a processed new payload. We've
		// already taken care of cases with only a comment or a
		// comment with a label
		newpl := ""
		cleanpl := strings.TrimSpace(j.payload)
		splitpl := strings.Fields(cleanpl)
		testpl := splitpl[0]

		switch {

		// At the moment, we do not translate directives but just pass them on
		// because there are so frickin' many variations out there
		// TODO convert .origin and .byte, .word etc payloads
		case isDirective(testpl):
			newpl = j.payload

		case isOpcode(testpl):
			oldmnem := Opcodes.Table[testpl].OldMnem
			comment := ""

			// TODO We include size data in the opcode list to be
			// able to validate the size of the operand, which is
			// important for tradition formats. This function is not
			// currently implemented
			// size := Opcodes.Table[testpl].Size

			if *upperOpcs {
				oldmnem = firstToUpper(oldmnem)
			}

			// Convert and insert operand.
			// TODO validate size of operad
			if strings.Contains(oldmnem, "?") {
				newop := convertNumber(splitpl[1])
				oldmnem = strings.Replace(oldmnem, "?", newop, -1)
			}

			// See if we have a comment after the instruction
			// TODO pretty format this
			if strings.Contains(cleanpl, ";") {
				comment = " ; " + strings.SplitN(cleanpl, ";", 2)[1]
			}

			newpl = opcodeindent + oldmnem + comment

		// Everything else is considered an error
		default:
			fmt.Println("FATAL: Unrecognized payload in line", j.linenumber)
			newpl = "ERROR -->" + j.payload
		}

		// Add any label back to line
		if label != "" {

			if *labelColon {
				label = label + ":"
			}

			labelindent := strings.Repeat(" ", len(opcodeindent)-len(label))
			newpl = label + labelindent + strings.TrimSpace(newpl)
		}

		j.payload = newpl

		results <- j
	}
}

// Basic structure of Worker Pools provided by
// https://gobyexample.com/worker-pools
func main() {

	// ---- SETUP ----

	flag.Parse()

	// IMPORT RAW SOURCE TEXT
	// It's a fatal error not to have a filename
	if *input == "" {
		log.Fatal("No input filename provided.")
	}

	inputFile, err := os.Open(*input)

	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	source := bufio.NewScanner(inputFile)

	for source.Scan() {
		raw = append(raw, source.Text())
	}

	// IMPORT OPCODES

	jfile, err := os.Open(OPCODELIST)
	if err != nil {
		log.Fatal(err)
	}

	jParser := json.NewDecoder(jfile)
	if err = jParser.Decode(&Opcodes); err != err {
		log.Fatal(err)
	}

	// ---- PROCESS LINES ----

	jobs := make(chan workline, len(raw))
	results := make(chan workline, len(raw))

	// Start as many jobs as we have CPU cores
	fmt.Println("Starting", maxworkers, "concurrent workers.")
	for w := 1; w <= maxworkers; w++ {
		go procLine(jobs, results)
	}

	// Send lines to the workers
	for n, l := range raw {
		jobs <- workline{n + 1, l}
	}
	close(jobs)

	// Get results back from workers
	for i := 1; i <= len(raw); i++ {
		processed = append(processed, <-results)
	}

	// ---- SORT RESULTS ----

	sort.Sort(byLine(processed))

	// TESTING TODO
	for _, l := range processed {
		fmt.Println(l.payload)
	}

}
