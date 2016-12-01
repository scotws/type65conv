# typ65conv - A Conversion Tool for Typist's Assembler to Traditional Formats

Scot W. Stevenson <scot.stevenson@gmail.com>

## Overview

Typist's Assembler Format (TAN) is an improved syntax for the 6502/65c02/65816
family of 8/16-bit processors. However, it is not widespread. To give coders who
(foolishly) insist on using the traditional format access to programs coded in
TAN, this tool converts it.

This is a BETA version. Use at your own risk.

For details on TAN, see the [introduction](https://docs.google.com/document/d/16Sv3Y-3rHPXyxT1J3zLBVq4reSPYtY2G6OSojNTm4SQ/).

## Usage

typ65conv is a command line tool. It requires an input file name, other
parameters are optional. To convert the file `example.tasm` to `old.asm` you
would use the line `typ65conv -i example.tasm -o old.asm.` This assumes that you
have Go installed and have compiled the source. To test, prefix the line above
with `go run`.


### Flags

**-i**  - Input file name (REQUIRED)
**-o**  - Output file name (default "typ65conv.asm")
**-lc** - Add colons to all labels (default don't). `label` -> `label:`
**-ou** - Make mnemonics uppercase (default lowercase). `nop` -> `NOP`


## Conversion

Currently, conversion is limited to mnemonics - opcodes and operands. Directives
are currently not touched because of the large number of variants in use. In
future, the parameters of certain directives such as `.origin` and `.byte`
should be converted. 


### Mnemonics

Mnemonics are converted from TAN to the traditional format, inserting operands
at the appropriate place (`lda.x 1000` becomes `lda $1000,x`). At the moment,
this is the main function of the tool. 


### Numbers

All numbers have the separators removed (usually `.` and `:`), turning `00:0000`
into `000000`. 

Decimal numbers have the `&` symbol removed (`&00` becomes `00`). Binary
numbers are kept unchanged. Octal numbers are not supported. Forced hexadecimal
numbers (`$` or `0x` prefix) are converted to normal hex (`$1000`). 

TAN assumes "naked" numbers such as `0000` to be hex to reduce visual clutter.
They are converted to traditional hexformat (`$0000`). This means that any
operand that can be a symbol or a hexadecimal number will be interpreted as a
number first (`dad` becomes `$dad`). 


### Labels

Labels can have a colon added, turning `label` into `label:`. There is
currently no automatic conversion to make them upper case, because that would
require the program to find labels in the text as well. This might be added in
the future.


## Technical Details

### Concurrency

Typ65conv makes use of Go's concurrency support to convert one line per logical
CPU core at a time. Note that for very small programs, this might actually make
the conversion process slower because of the overhead of creating goroutines.
