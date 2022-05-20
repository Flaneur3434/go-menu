package main

import (
	"9fans.net/go/draw"
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"
)

//  dmenu  [-bfiv]  [-l  lines]  [-m monitor] [-p prompt] [-fn font] [-nb color]
//       [-nf color] [-sb color] [-sf color] [-nhb color] [-nhf color]  [-shb  color]
//       [-shf color] [-w windowid]

var (
	bottom          bool
	grabKeyBoard    bool
	caseInsensitive bool
	lines           int
	monitor         int
	prompt          string
	font            string
	normBackg       string
	normForeg       string
	selBackg        string
	selForeg        string
	normHighlBack   string
	normHighlFore   string
	selHighlBack    string
	selHighlFore    string
	version         bool
	windowId        string
	help            bool
	display         *draw.Display
	keyboard        *draw.Keyboardctl
)

func init() {
	flag.BoolVar(&bottom, "b", false, "dmenu appears at the bottom of the screen")
	flag.BoolVar(&grabKeyBoard, "f", false, "dmenu  grabs  the keyboard before reading stdin if not reading from a tty. This  is  faster,  but  will  lock  up  X  until  stdin  reaches end-of-file.")
	flag.BoolVar(&caseInsensitive, "i", false, "dmenu matches menu items case insensitively")
	flag.IntVar(&lines, "l", 5, "dmenu lists items vertically, with the given number of lines")
	flag.IntVar(&monitor, "m", 0, "dmenu  is  displayed  on the monitor number supplied. Monitor numbers are starting from 0")
	flag.StringVar(&prompt, "p", "", "defines the prompt to be displayed to the left of the input field")
	flag.StringVar(&font, "fn", "", "defines the font or font set used")
	flag.StringVar(&normBackg, "nb", "", "defines the normal background color. #RGB, #RRGGBB, and X color names are supported")
	flag.StringVar(&normForeg, "nf", "", "defines the normal foreground color")
	flag.StringVar(&selBackg, "sb", "", "defines the selected background color")
	flag.StringVar(&selForeg, "sf", "", "defines the selected foreground color")
	flag.StringVar(&normHighlBack, "nhb", "", "defines the normal highlight background color")
	flag.StringVar(&normHighlFore, "nhf", "", "defines the normal highlight foreground color")
	flag.StringVar(&selHighlBack, "shb", "", "defines the selected highlight background color")
	flag.StringVar(&selHighlFore, "shf", "", "defines the selected highlight foreground color")
	flag.BoolVar(&version, "v", false, "prints version information to stdout, then exits")
	flag.StringVar(&windowId, "w", "", "embed into windowid")
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

	display, _ = draw.Init(nil, "", "test", "")
	keyboard = display.InitKeyboard()

	scanner := bufio.NewScanner(os.Stdin)
	const inputDefaultSize = 4098
	input := make([]byte, inputDefaultSize)

	// store stdin into 'input' slice
	for scanner.Scan() {
		input = append(input, scanner.Bytes()...)
		input = append(input, '\n')
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading err: ", err)
	}

	display.AllocImage(draw.Rect(100, 100, 200, 200), draw.RGB24, true, draw.Red)
	display.Flush()

	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("%s", input)
}
