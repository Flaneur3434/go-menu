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
	defaultWinSizeH           = 480
	defaultWinSizeW           = 640
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
	numOfRows  int
	itemList   []string
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
	menuChan := make(chan *Menu)
	errChan := make(chan error)
	input := make([]string, inputDefaultSize)
	var menu *Menu

	go setUpMenu(menuChan, errChan)
	go readInput(inputChan)

	menu = <-menuChan
	if err := <-errChan; err != nil {
		panic(err)
	}

	numOfInput := 0
	for s := range inputChan {
		if numOfInput+1 >= len(input) {
			input = append(input, make([]string, int(float64(cap(input))*defaultGrowthSize))...)
		}
		input[numOfInput] = s
		numOfInput += 1
	}
	menu.numOfItems = numOfInput
	menu.itemList = input

	if err := menu.writeItem(); err != nil {
		panic(err)
	}

	// main loop
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				readKey(t)
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

func setUpMenu(menuChan chan *Menu, errChan chan error) {
	m := Menu{window: nil, surface: nil, font: nil, numOfItems: 0, numOfRows: int(defaultWinSizeH / fontSize), itemList: nil}
	var err error

	// create window
	pixelHeight := int(defaultWinSizeH/fontSize)*fontSize + int(defaultWinSizeH/fontSize)
	m.window, err = sdl.CreateWindow("go-menu", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, defaultWinSizeW, int32(pixelHeight), sdl.WINDOW_SHOWN|sdl.WINDOW_SKIP_TASKBAR)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		menuChan <- nil
		errChan <- err
		return
	}

	// get surface
	if m.surface, err = m.window.GetSurface(); err != nil {
		menuChan <- nil
		errChan <- err
		return
	}

	// get font
	if m.font, err = ttf.OpenFont(fontPath, fontSize); err != nil {
		menuChan <- nil
		errChan <- err
		return
	}

	menuChan <- &m
	errChan <- nil
}

func (m *Menu) writeItem() (err error) {
	// get text
	var text *sdl.Surface
	renderTextSlice := make([]*sdl.Surface, m.numOfRows)

	for i := 0; i < len(renderTextSlice); i++ {
		text, err = m.font.RenderUTF8Blended(m.itemList[i], sdl.Color{R: 255, G: 0, B: 0, A: 255})
		if err != nil {
			return err
		}

		renderTextSlice[i] = text
	}

	for i := 0; i < len(renderTextSlice); i++ {
		if err = renderTextSlice[i].Blit(nil, m.surface, &sdl.Rect{X: 1, Y: 1 + int32((i * fontSize)), W: 0, H: 0}); err != nil {
			return
		}
	}

	m.window.UpdateSurface()

	for _, sur := range renderTextSlice {
		defer sur.Free()
	}
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

	os.Exit(0)
}

func readKey(keyPress *sdl.KeyboardEvent) {
	keyCode := keyPress.Keysym.Sym
	keys := ""

	// Modifier keys
	switch keyPress.Keysym.Mod {
	case sdl.KMOD_LALT:
		keys += "Left Alt"
	case sdl.KMOD_LCTRL:
		keys += "Left Control"
	case sdl.KMOD_LSHIFT:
		keys += "Left Shift"
	case sdl.KMOD_LGUI:
		keys += "Left Meta or Windows key"
	case sdl.KMOD_RALT:
		keys += "Right Alt"
	case sdl.KMOD_RCTRL:
		keys += "Right Control"
	case sdl.KMOD_RSHIFT:
		keys += "Right Shift"
	case sdl.KMOD_RGUI:
		keys += "Right Meta or Windows key"
	case sdl.KMOD_NUM:
		keys += "Num Lock"
	case sdl.KMOD_CAPS:
		keys += "Caps Lock"
	case sdl.KMOD_MODE:
		keys += "AltGr Key"
	}

	if keyCode < 10000 {
		if keys != "" {
			keys += " + "
		}

		// If the key is held down, this will fire
		if keyPress.Repeat > 0 {
			keys += string(keyCode) + " repeating"
		} else {
			switch keyPress.State {
			case sdl.RELEASED:
				keys += string(keyCode) + " released"
			case sdl.PRESSED:
				keys += string(keyCode) + " pressed"
			}
		}

	}

	if keys != "" {
		fmt.Println(keys)
	}
}
