# typ65conv - A Conversion Tool for Typist's Assembler to Traditional Formats

Scot W. Stevenson <scot.stevenson@gmail.com>

## Overview

Typist's Assembler Format (TAN) is an improved syntax for the 6502/65c02/65816
family of 8/16-bit processors. However, it is not widespread. To give coders who
(foolishly) insist on using the traditional format access to programs coded in
TAN, this tool converts it.

## Usage

### Flags

**-i**  - Input file name (REQUIRED)
**-o**  - Output file name (default "typ65conv.asm")
**-lc** - Add colons to all labels (default don't). `label` -> `label:`
**-ou** - Make opcodes uppercase (default lowercase). `nop` -> `NOP`


## Conversion


### Numbers

Numbers have the separators removed (usually `.` and `:`), turning `00:0000`
into `000000`.


### Labels

Labels can automatically have a colon added, turning `label` into `label:`.
There is currently no automatic conversion to make them upper case, because 
that would require the program to find labels in the text as well. This
might be added in the future.


## Technical Details

### Concurrency

Typ65conv makes use of Go's concurrency support to convert one line per logical
CPU core at a time. Note that for very small programs, this might actually make
the conversion process slower because of the overhead of creating goroutines.


