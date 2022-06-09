package main

import (
	"flag"
	"fmt"
	"strings"

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
	flag.StringVar(&fontPath, "fn", "./assets/zpix.ttf", "defines the font or font set used")
	flag.StringVar(&normBackg, "nb", "#cccccc", "defines the normal background color. #RGB, #RRGGBB, and X color names are supported")
	flag.StringVar(&normForeg, "nf", "#000000", "defines the normal foreground color")
	flag.StringVar(&selBackg, "sb", "#0066ff", "defines the selected background color")
	flag.StringVar(&selForeg, "sf", "#ffffff", "defines the selected foreground color")
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

	if help {
		flag.VisitAll(func(flag *flag.Flag) {
			format := "\t-%s: %s (Default: '%s')\n"
			fmt.Printf(format, flag.Name, flag.Usage, flag.DefValue)
		})
		return
	}

	stdInChan := make(chan string)
	input := make([]string, 0, inputDefaultSize)
	go util.ReadStdIn(stdInChan)

	menu, err := draw.SetUpMenu(fontPath, normBackg, normForeg, selBackg, selForeg)
	if err != nil {
		panic(err)
	}

	// store results from ReadStdIn
	for s := range stdInChan {
		if len(input) == cap(input) {
			input = append(input, make([]string, 0, int(float64(cap(input))*defaultGrowthSize))...)
		}
		input = append(input, s)
	}

	// draw the initial view of the gui from ReadStdIn
	fuzzList := util.InitRanks(input)
	menu.SetNumOfItem(fuzzList.Len())
	if err := menu.WriteItem(&fuzzList); err != nil {
		panic(err)
	}

	// main loop
	running := true
	keyBoardChan := make(chan string)
	updateChan := make(chan bool)
	newRanksChan := make(chan util.Ranks)
	prevRanks := fuzzList

	for running {
		event := sdl.PollEvent()
		switch t := event.(type) {
		case *sdl.QuitEvent:
			running = false
		case *sdl.TextInputEvent:
			go func() {
				if caseInsensitive {
					keyBoardInput += strings.ToLower(t.GetText())
				} else {
					keyBoardInput += t.GetText()
				}
				menu.WriteKeyBoard(keyBoardInput)
				keyBoardChan <- keyBoardInput
			}()
		case *sdl.KeyboardEvent:
			// TODO: unicode support, shift key
			if t.State == sdl.PRESSED {
				switch t.Keysym.Sym {
				case sdl.K_BACKSPACE:
					if len(keyBoardInput) > 0 {
						keyBoardInput = keyBoardInput[:len(keyBoardInput)-1]
						menu.WriteKeyBoard(keyBoardInput)
						keyBoardChan <- keyBoardInput
					}
				case sdl.K_RETURN:
					menu.GetSelItem(&prevRanks)
					running = false
				case sdl.K_UP:
					menu.ScrollMenuUp(&prevRanks)
					updateChan <- true
				case sdl.K_DOWN:
					menu.ScrollMenuDown(&prevRanks)
					updateChan <- true
				}
			}
		}

		go func() {
			select {
			case keyBoardInput := <-keyBoardChan:
				newRanksChan <- util.FuzzySearch(input, keyBoardInput, caseInsensitive)
			case <-updateChan:
				// need to synce with the first go routine
				updateChan <- true
			}
		}()

		select {
		case ranks := <-newRanksChan:
			menu.ResetPosCounters()
			// sometimes the screen overlaps with prev
			menu.WriteItem(&ranks)
			// select statement acts like a mutex and only one thread can update this variable at a time
			prevRanks = ranks
		case <-updateChan:
			// show previous fuzzList to screen
			menu.WriteItem(&prevRanks)
		default:
		}

		sdl.Delay(10)
	}

	menu.CleanUp()
}

// TODO: use the line argument
