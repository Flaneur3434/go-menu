package main

import (
	"flag"
	"fmt"

	"github.com/Flaneur3434/go-menu/draw"
	"github.com/Flaneur3434/go-menu/util"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

//  dmenu  [-bfiv]  [-l  lines]  [-m monitor] [-p prompt] [-fn font] [-nb color]
//       [-nf color] [-sb color] [-sf color] [-nhb color] [-nhf color]  [-shb  color]
//       [-shf color] [-w windowid]

const (
	inputDefaultSize          = 10
	defaultGrowthSize float64 = 1.2
)

var (
	bottom          bool
	grabKeyBoard    bool
	caseInsensitive bool
	lines           int
	monitor         int
	prompt          string
	fontPath        string
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
)

func init() {
	flag.BoolVar(&bottom, "b", false, "dmenu appears at the bottom of the screen")
	flag.BoolVar(&grabKeyBoard, "f", false, "dmenu  grabs  the keyboard before reading stdin if not reading from a tty. This  is  faster,  but  will  lock  up  X  until  stdin  reaches end-of-file.")
	flag.BoolVar(&caseInsensitive, "i", false, "dmenu matches menu items case insensitively")
	flag.IntVar(&lines, "l", 5, "dmenu lists items vertically, with the given number of lines")
	flag.IntVar(&monitor, "m", 0, "dmenu  is  displayed  on the monitor number supplied. Monitor numbers are starting from 0")
	flag.StringVar(&prompt, "p", "", "defines the prompt to be displayed to the left of the input field")
	flag.StringVar(&fontPath, "fn", "", "defines the font or font set used")
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

	// init fonts
	if err := ttf.Init(); err != nil {
		panic(err)
	}

	// init sdl
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}

	// default font
	if fontPath == "" {
		fontPath = "./assets/zpix.ttf"
	}

	// default number of lines
	if lines == 0 {
		lines = 4
	}

	if help {
		flag.VisitAll(func(flag *flag.Flag) {
			format := "\t-%s: %s (Default: '%s')\n"
			fmt.Printf(format, flag.Name, flag.Usage, flag.DefValue)
		})
	}

	stdInChan := make(chan string)
	menuChan := make(chan *draw.Menu)
	errChan := make(chan error)
	keyBoardChan := make(chan string)
	ranksChan := make(chan util.Ranks)

	input := make([]string, 0, inputDefaultSize)
	var menu *draw.Menu

	go draw.SetUpMenu(fontPath, menuChan, errChan)
	go util.ReadStdIn(stdInChan)

	for s := range stdInChan {
		if len(input) == cap(input) {
			input = append(input, make([]string, 0, int(float64(cap(input))*defaultGrowthSize))...)
		}
		input = append(input, s)
	}

	menu = <-menuChan
	if err := <-errChan; err != nil {
		panic(err)
	}

	menu.ItemList = input
	fuzzList := util.InitRanks(menu.ItemList)

	if err := menu.WriteItem(fuzzList); err != nil {
		panic(err)
	}

	if err := menu.WriteKeyBoard(); err != nil {
		panic(err)
	}

	// main loop
	running := true
	for running {
		event := sdl.PollEvent()
		switch t := event.(type) {
		case *sdl.QuitEvent:
			running = false
		case *sdl.KeyboardEvent:
			go func() {
				// TODO: unicode support
				if t.State == sdl.RELEASED {
					switch t.Keysym.Sym {
					case 8:
						if len(menu.KeyBoardInput) > 0 {
							menu.KeyBoardInput = menu.KeyBoardInput[:len(menu.KeyBoardInput)-1]
						}
					default:
						menu.KeyBoardInput += string(t.Keysym.Sym)
					}
				}
				keyBoardChan <- menu.KeyBoardInput
			}()

			go func() {
				util.FuzzySearch(&fuzzList, <-keyBoardChan)
				ranksChan <- fuzzList
			}()

			select {
			case ranks := <-ranksChan:
				menu.WriteItem(ranks)
				menu.WriteKeyBoard()
			default:
				menu.WriteKeyBoard()
			}
		}
		sdl.Delay(3)
	}

	// TODO: os.stdout the selected item
	menu.CleanUp()
}
