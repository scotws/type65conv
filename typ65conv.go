// typ65conv - Convert Typist's Assembler Code to Traditional Formats
// Scot W. Stevenson <scot.stevenson@gmail.com>
// First version: 16. September 2016
// This version: 16. November 2016

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
	"strings"
)

// We keep line numbers associated with each line while we work on things for
// the error codes and to be able to stor them as the last step
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
	upperDirs  = flag.Bool("du", false, "Convert directives to upper case")
	labelColon = flag.Bool("lc", false, "Add colon to labels")
	input      = flag.String("i", "", "Input file (REQUIRED)")
	output     = flag.String("o", "typ65conv.asm", "Output file (default 'typ65conv.asm')")
)

// Defintions for sorting, see https://golang.org/pkg/sort/
// We sort by line numbers
type byLine []workline

func (a byLine) Len() int           { return len(a) }
func (a byLine) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLine) Less(i, j int) bool { return a[i].linenumber < a[j].linenumber }

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
// Typist Assembler Notation opcodes, it returns the bool true, else false
// TODO write test routine
func isOpcode(s string) bool {
	// TODO write code
	return true
}

// mergeLabel takes to strings, the first a complete line with the original
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
// in the string, the original version is returned
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

		if hasLabel(j.payload) {

			// Assumes we have already checked for empty lines
			labelline := strings.Fields(j.payload)
			label = labelline[0]

			if *labelColon {
				label = label + ":"
			}

			// If there is only the label in the line, we can deal
			// with it immediately
			if len(labelline) == 1 {
				results <- workline{j.linenumber, label}
				continue
			}

			// If there is only the label and a comment, we can
			// reassemble the line right now and continue
			// TODO make sure there is enough whitespace left in the
			// cases where the label gets a colon added
			if isComment(labelline[1]) {
				results <- workline{j.linenumber, mergeLabel(j.payload, label)}
				continue
			}

			// There is something else on the line we have to
			// process first, then we can come back and reassemble
			// stuff. We use the fact that label is not empty as a
			// flag
			j.payload = labelline[1]

		}

		// Main switch statement returns a processed new payload
		newpayload := ""

		switch {

		case isDirective(j.payload):

			// See if we need to convert the first word, that is,
			// the directive itself, to uppercase

			// TODO take care of numbers

			if *upperDirs {
				newpayload = firstToUpper(j.payload)
			} else {
				newpayload = j.payload
			}

		case isComment(j.payload):

			newpayload = j.payload

		case isOpcode(j.payload):

			if *upperOpcs {
				newpayload = firstToUpper(j.payload)
			} else {
				newpayload = j.payload

			}

			// TODO convert instruction
			// TODO handle in-line comments
			newpayload = j.payload

		// Everything else is considered an error
		default:
			fmt.Println("FATAL: Unrecognized payload in line", j.linenumber)
			newpayload = "ERROR -->" + j.payload
		}

		// Add any label back to line
		if label != "" {
			newpayload = label + " " + newpayload
		}

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

	// TODO Testing only, remove when done
	fmt.Println(Opcodes)

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
