package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

//  dmenu  [-bfiv]  [-l  lines]  [-m monitor] [-p prompt] [-fn font] [-nb color]
//       [-nf color] [-sb color] [-sf color] [-nhb color] [-nhf color]  [-shb  color]
//       [-shf color] [-w windowid]

const (
	fontSize                  = 16
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

type Menu struct {
	window     *sdl.Window
	surface    *sdl.Surface
	font       *ttf.Font
	numOfItems int
	itemList   []Item
}

type Item struct {
	renderedText *sdl.Surface
	text         string
}

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

	inputChan := make(chan string)
	input := make([]string, inputDefaultSize)

	go readInput(inputChan)

	numOfInput := 0
	for s := range inputChan {
		if numOfInput+1 >= len(input) {
			input = append(input, make([]string, int(float64(cap(input))*defaultGrowthSize))...)
		}
		input[numOfInput] = s
		numOfInput += 1
	}

	menu, err := setUpMenu(input, numOfInput)
	if err != nil {
		panic(err)
	}

	err = menu.writeItem()
	if err != nil {
		panic(err)
	}

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}

		sdl.Delay(16)
	}

	menu.cleanUp()
}

func readInput(c chan string) {
	scanner := bufio.NewScanner(os.Stdin)

	// store stdin into 'input' slice
	for scanner.Scan() {
		c <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading err: ", err)
	}

	close(c)
}

func setUpMenu(input []string, numOfItems int) (*Menu, error) {
	var err error
	var text *sdl.Surface

	m := Menu{window: nil, surface: nil, font: nil, numOfItems: numOfItems, itemList: make([]Item, numOfItems)}

	// create window
	m.window, err = sdl.CreateWindow("go-menu", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 640, 480, sdl.WINDOW_SHOWN|sdl.WINDOW_BORDERLESS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return nil, err
	}

	// get surface
	if m.surface, err = m.window.GetSurface(); err != nil {
		return nil, err
	}

	// get font
	if m.font, err = ttf.OpenFont(fontPath, fontSize); err != nil {
		return nil, err
	}

	// get text
	for i := 0; i < numOfItems; i++ {
		text, err = m.font.RenderUTF8Blended(input[i], sdl.Color{R: 255, G: 0, B: 0, A: 255})
		if err != nil {
			return nil, err
		}

		if i+1 >= len(m.itemList) {
			m.itemList = append(m.itemList, make([]Item, int(float64(cap(m.itemList))*defaultGrowthSize))...)
		}

		m.itemList[i] = Item{renderedText: text, text: input[i]}
	}

	return &m, nil
}

func (m *Menu) writeItem() (err error) {
	for i := 0; i < m.numOfItems && i < int(m.surface.H/fontSize); i++ {
		if err = m.itemList[i].renderedText.Blit(nil, m.surface, &sdl.Rect{X: 1, Y: 1 + int32((i * fontSize)), W: 0, H: 0}); err != nil {
			return
		}
	}

	m.window.UpdateSurface()

	return
}

func (m *Menu) cleanUp() {
	ttf.Quit()
	sdl.Quit()

	if m.window != nil {
		defer m.window.Destroy()
	}

	if m.font != nil {
		defer m.font.Close()
	}

	for i := range m.itemList {
		defer m.itemList[i].renderedText.Free()
	}

	os.Exit(0)
}
