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
	// flag variables
	grabKeyBoard    bool
	caseInsensitive bool
	lines           int
	prompt          string
	fontPath        string
	normBackg       string
	normForeg       string
	selBackg        string
	selForeg        string
	help            bool

	// variables used in main function
	keyBoardInput string
	menu          *draw.Menu
)

func init() {
	flag.BoolVar(&grabKeyBoard, "f", false, "dmenu  grabs  the keyboard before reading stdin if not reading from a tty. This  is  faster,  but  will  lock  up  X  until  stdin  reaches end-of-file.")
	flag.BoolVar(&caseInsensitive, "i", false, "dmenu matches menu items case insensitively")
	flag.IntVar(&lines, "l", 5, "dmenu lists items vertically, with the given number of lines")
	flag.StringVar(&prompt, "p", "", "defines the prompt to be displayed to the left of the input field")
	flag.StringVar(&fontPath, "fn", "", "defines the font or font set used")
	flag.StringVar(&normBackg, "nb", "", "defines the normal background color. #RGB, #RRGGBB, and X color names are supported")
	flag.StringVar(&normForeg, "nf", "", "defines the normal foreground color")
	flag.StringVar(&selBackg, "sb", "", "defines the selected background color")
	flag.StringVar(&selForeg, "sf", "", "defines the selected foreground color")
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
	go util.ReadStdIn(stdInChan)

	input := make([]string, 0, inputDefaultSize)
	menuChan := make(chan *draw.Menu)
	errChan := make(chan error)
	go draw.SetUpMenu(fontPath, input, menuChan, errChan, normBackg, normForeg, selBackg, selForeg)

	// store results from ReadStdIn
	for s := range stdInChan {
		if len(input) == cap(input) {
			input = append(input, make([]string, 0, int(float64(cap(input))*defaultGrowthSize))...)
		}
		input = append(input, s)
	}

	// handle any errors from SetUpMenu
	menu = <-menuChan
	if err := <-errChan; err != nil {
		panic(err)
	}

	// draw the initial view of the gui from ReadStdIn
	fuzzList := util.InitRanks(input)
	if err := menu.WriteItem(fuzzList); err != nil {
		panic(err)
	}

	if err := menu.WriteKeyBoard(keyBoardInput); err != nil {
		panic(err)
	}

	// main loop
	running := true
	keyBoardChan := make(chan string)
	updateChan := make(chan bool)
	ranksChan := make(chan util.Ranks)

	for running {
		event := sdl.PollEvent()
		switch t := event.(type) {
		case *sdl.QuitEvent:
			running = false
		case *sdl.KeyboardEvent:
			go func() {
				// TODO: unicode support, shift key
				if t.State == sdl.PRESSED {
					switch t.Keysym.Sym {
					case sdl.K_BACKSPACE:
						if len(keyBoardInput) > 0 {
							keyBoardInput = keyBoardInput[:len(keyBoardInput)-1]
							keyBoardChan <- keyBoardInput
						}
					case sdl.K_UP:
						menu.ScrollMenuUp()
						updateChan <- true
					case sdl.K_DOWN:
						menu.ScrollMenuDown()
						updateChan <- true
					default:
						keyBoardInput += string(t.Keysym.Sym)
						keyBoardChan <- keyBoardInput
					}
				}
				menu.WriteKeyBoard(keyBoardInput)
			}()
		}

		go func() {
			select {
			case keyBoardInput := <-keyBoardChan:
				ranksChan <- util.FuzzySearch(input, keyBoardInput)
			case <-updateChan:
				// need to synce with the first go routine
				updateChan <- true
			}
		}()

		select {
		case ranks := <-ranksChan:
			menu.ResetTopItem()
			// sometimes the screen overlaps with prev
			menu.WriteItem(ranks)
		case <-updateChan:
			// show previous fuzzList to screen
			menu.WriteItem(fuzzList)
		default:
		}

		sdl.Delay(3)
	}

	menu.CleanUp()
}
