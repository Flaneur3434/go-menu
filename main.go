package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

//  dmenu  [-bfiv]  [-l  lines]  [-m monitor] [-p prompt] [-fn font] [-nb color]
//       [-nf color] [-sb color] [-sf color] [-nhb color] [-nhf color]  [-shb  color]
//       [-shf color] [-w windowid]

var bottom *bool = flag.Bool("b", false, "dmenu appears at the bottom of the screen")
var grabKeyBoard *bool = flag.Bool("f", false, "dmenu  grabs  the keyboard before reading stdin if not reading from a tty. This  is  faster,  but  will  lock  up  X  until  stdin  reaches end-of-file.")
var caseInsensitive *bool = flag.Bool("i", false, "dmenu matches menu items case insensitively")
var lines *int = flag.Int("l", 5, "dmenu lists items vertically, with the given number of lines")
var monitor *int = flag.Int("m", 0, "dmenu  is  displayed  on the monitor number supplied. Monitor numbers are starting from 0")
var prompt *string = flag.String("p", "", "defines the prompt to be displayed to the left of the input field")
var font *string = flag.String("fn", "", "defines the font or font set used")
var normBackg *string = flag.String("nb", "", "defines the normal background color.   #RGB,  #RRGGBB,  and  X  color names are supported")
var normForeg *string = flag.String("nf", "", "defines the normal foreground color")
var selBackg *string = flag.String("sb", "", "defines the selected background color")
var selForeg *string = flag.String("sf", "", "defines the selected foreground color")
var normHighlBack *string = flag.String("nhb", "", "defines the normal highlight background color")
var normHighlFore *string = flag.String("nhf", "", "defines the normal highlight foreground color")
var selHighlBack *string = flag.String("shb", "", "defines the selected highlight background color")
var selHighlFore *string = flag.String("shf", "", "defines the selected highlight foreground color")
var version *bool = flag.Bool("v", false, "prints version information to stdout, then exits")
var windowId *string = flag.String("w", "", "embed into windowid")
var help bool

func init() {
	flag.BoolVar(&help, "h", false, "print help message")
	flag.BoolVar(&help, "help", false, "print help message")
}

func main() {
	flag.Parse()

	if help {
		flag.VisitAll(func(flag *flag.Flag) {
			format := "\t-%s: %s (Default: '%s')\n"
			fmt.Printf(format, flag.Name, flag.Usage, flag.DefValue)
		})
	}

	scanner := bufio.NewScanner(os.Stdin)
	const inputDefaultSize = 4098
	input := make([]byte, inputDefaultSize)

	// store stdin into 'input' slice
	for scanner.Scan() {
		input = append(input, scanner.Bytes()...)
		// input = append(input, '\n')
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading err: ", err)
	}

	fmt.Printf("%s", input)
}
