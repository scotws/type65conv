# Convert list of 65c02/65816 opcodes to JSON format 
# Scot W. Stevenson <scot.stevenson@gmail.com>
# First version: 30. Nov 2016
# This version: 30. Nov 2016

# Tool to convert list of Typist Assembler Mnemonics in the form
#
#   brk 2
#
# to the JSON format used by typ65conv.go (see https://github.com/scotws/type65conv)

import sys

if sys.version_info.major != 3:
    print("FATAL: Python 3 required. Aborting.")
    sys.exit(1)


# CONSTANTS

SOURCEFILE = "opcodes65c02+65816.txt"
DESTFILE = "opcodes.json"

HEADER = """{
        "table": {\n"""

FOOTER = """        }
}\n"""

INDENT1 = "                "
INDENT2 = INDENT1 + "        "


# TABLES

tails = {\
        "" : "?",
        "#": "#?",
        "a": "a",
        "d": "?",
        "di": "(?)",
        "dil": "[?]",
        "dily": "[?],y",
        "diy": "(?),y",
        "dx": "?,x",
        "dy": "?,y",
        "dxi": "(?,x)",
        "i": "(?)",
        "il": "[?]",
        "l": "?",
        "lx": "?,x",
        "r": "?",
        "s": "?,s",
        "siy": "(?,s),y",
        "x": "?,x",
        "xi": "(?,x)",
        "y": "?,y",
        "z": "?",
        "zi": "(?)",
        "ziy": "(?),y",
        "zx": "?,x",
        "zxi": "(?,x)",
        "zy": "?,y",
        }

specialcases = {\
        "bra.l": "brl",
        "jmp.l": "jml",
        "jsr.l": "jsl",
        "phe.#": "pea",
        "phe.d": "pei",
        "phe.r": "per",
        "rts.l": "rtl",
        }

with open(SOURCEFILE, "r") as f:
    src = f.readlines()

output = []

for l in src:
    m, s = l.split()

    try:
        b, t = m.split(".")
    except ValueError:
        b = m
        t = ""

    if m in specialcases:
        b = specialcases[m]

    # Need space if there is an operand
    # Only lookup tails if size > 1
    if (int(s) > 1) or (t == "a"):
        b = b + " "
        t = tails[t]

    # phe.# is a really, really special case
    if m == "phe.#":
        t = "?"

    # So is rts.l because it is one byte long (RTL)
    if m == "rts.l":
        t = ""

    # Keep the following lines in source code for testing
#   print('{0}"{1}": {{'.format(INDENT1, m))
#   print('{0}"oldmnem": "{1}{2}",'.format(INDENT2, b, t))
#   print('{0}"size": {1}}},'.format(INDENT2, s))

    output.append('{0}"{1}": {{'.format(INDENT1, m))
    output.append('{0}"oldmnem": "{1}{2}",'.format(INDENT2, b, t))
    output.append('{0}"size": {1}}},'.format(INDENT2, s))


with open(DESTFILE, "w") as d:
    d.write(HEADER)

    for l in output:
        d.write(l+"\n")

    d.write(FOOTER)

print("Please remove last comma from last line of list manually")

