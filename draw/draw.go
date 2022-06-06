package draw

import (
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/Flaneur3434/go-menu/util"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	fontSize        = 16
	defaultWinSizeH = 480
	defaultWinSizeW = 640
)

var (
	mu          sync.Mutex
	currentItem int = 0
	topItem     int = 0
)

type Menu struct {
	window    *sdl.Window
	renderer  *sdl.Renderer
	surface   *sdl.Surface
	font      *ttf.Font
	numOfRows int
	numOfItem int

	normBackg string
	normForeg string
	selBackg  string
	selForeg  string
}

func SetUpMenu(fontPath string, menuChan chan *Menu, errChan chan error, normBackg, normForeg, selBackg, selForeg string) {
	var err error
	m := Menu{window: nil,
		surface:   nil,
		font:      nil,
		numOfRows: int(defaultWinSizeH / fontSize),
		numOfItem: 0,
		normBackg: normBackg,
		normForeg: normForeg,
		selBackg:  selBackg,
		selForeg:  selForeg}

	// create window
	pixelHeight := int(defaultWinSizeH/fontSize)*fontSize + int(defaultWinSizeH/fontSize) + fontSize
	m.window, err = sdl.CreateWindow("go-menu", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, defaultWinSizeW, int32(pixelHeight), sdl.WINDOW_SHOWN|sdl.WINDOW_SKIP_TASKBAR)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		menuChan <- nil
		errChan <- err
		return
	}

	// create renderer
	m.renderer, err = sdl.CreateRenderer(m.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
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

func (m *Menu) WriteItem(R *util.Ranks) error {
	// clear clear of any artifacts
	m.renderer.Clear()
	m.surface.FillRect(&sdl.Rect{X: 0, Y: 0, W: defaultWinSizeW, H: defaultWinSizeH + 2}, 0)

	// probably can make parallel but drawNorm clears screen
	m.drawNorm(R)
	m.drawSel((*R)[currentItem+topItem].Word)

	m.window.UpdateSurface()
	return nil
}

func (m *Menu) WriteKeyBoard(keyBoardInput string) error {
	// clear clear of any artifacts
	m.renderer.Clear()
	m.surface.FillRect(&sdl.Rect{X: 1, Y: (defaultWinSizeH/fontSize)*fontSize + defaultWinSizeH/fontSize, W: defaultWinSizeW, H: fontSize}, 0)

	if len(keyBoardInput) > 0 {
		// render keyboard input
		text, err := m.font.RenderUTF8Blended(keyBoardInput, sdl.Color{R: 0, G: 255, B: 0, A: 255})
		if err != nil {
			return err
		}

		// draw keyboard input
		err = text.Blit(nil, m.surface, &sdl.Rect{X: 1, Y: (defaultWinSizeH/fontSize)*fontSize + defaultWinSizeH/fontSize, W: 0, H: 0})
		if err != nil {
			return err
		}

		defer text.Free()
	}

	m.window.UpdateSurface()
	return nil
}

func (m *Menu) CleanUp() {
	ttf.Quit()
	sdl.Quit()

	if m.window != nil {
		defer m.window.Destroy()
	}

	if m.window != nil {
		defer m.renderer.Destroy()
	}

	if m.font != nil {
		defer m.font.Close()
	}

	os.Exit(0)
}

func (m *Menu) SetNumOfItem(len int) {
	m.numOfItem = len
}

/*
 * currentItem can be in the range of 0 to 31
 * topItem is a multiple of m.numOfRows and can be 0 to (m.numOfItem - 1)
 */

// Move the menu down by one item
func (m *Menu) ScrollMenuDown() {
	mu.Lock()
	defer mu.Unlock()
	if topItem < m.numOfItem && (currentItem+1)%m.numOfRows == 0 {
		topItem += m.numOfRows
		currentItem = 0
	} else if (currentItem+1)%m.numOfRows != 0 {
		currentItem++
	}
}

// Move the menu Up by one item
func (m *Menu) ScrollMenuUp() {
	mu.Lock()
	defer mu.Unlock()
	if topItem >= m.numOfRows && currentItem%m.numOfRows == 0 {
		topItem -= m.numOfRows
		currentItem = m.numOfRows - 1
	} else if currentItem%m.numOfRows != 0 {
		currentItem--
	}
}

func (m *Menu) ResetPosCounters() {
	mu.Lock()
	defer mu.Unlock()
	topItem = 0
	currentItem = 0
}

// TODO: error checking, commenting, refactoring
func (m *Menu) drawSel(text string) (err error) {
	// using variables cause its easier to read
	var x1 int32 = 0
	var y1 int32 = int32(fontSize * currentItem)
	var x2 int32 = defaultWinSizeW
	var y2 int32 = y1 + fontSize

	var textRender *sdl.Surface
	var backGRender *sdl.Surface
	backGRender, err = sdl.CreateRGBSurface(0, defaultWinSizeW, fontSize, 32, 0, 0, 0, 0)

	// get color
	rB, gB, bB := util.ConvertStrToInt32(m.selBackg)
	rF, gF, bF := util.ConvertStrToInt32(m.selForeg)
	colorB := sdl.MapRGB(m.surface.Format, rB, gB, bB)

	// rendering part
	backGRender.FillRect(&sdl.Rect{X: 0, Y: 0, W: defaultWinSizeW, H: fontSize}, colorB)
	textRender, err = m.font.RenderUTF8Blended(text, sdl.Color{R: rF, G: gF, B: bF, A: 255})
	textRender.Blit(nil, backGRender, &sdl.Rect{X: 0, Y: 0, W: defaultWinSizeH, H: fontSize})
	backGRender.Blit(nil, m.surface, &sdl.Rect{X: x1, Y: y1, W: x2, H: y2})
	defer textRender.Free()
	defer backGRender.Free()

	return
}

// TODO: error checking, commenting, refactoring
func (m *Menu) drawNorm(R *util.Ranks) (err error) {
	var numRender int
	if len(*R) < m.numOfRows {
		numRender = len(*R)
	} else {
		numRender = m.numOfRows
	}
	renderTextSlice := make([]*sdl.Surface, numRender)

	var backGRender *sdl.Surface
	backGRender, err = sdl.CreateRGBSurface(0, defaultWinSizeW, defaultWinSizeH, 32, 0, 0, 0, 0)

	// get color
	rB, gB, bB := util.ConvertStrToInt32(m.normBackg)
	rF, gF, bF := util.ConvertStrToInt32(m.normForeg)

	colorB := sdl.MapRGB(m.surface.Format, rB, gB, bB)

	backGRender.FillRect(&sdl.Rect{X: 0, Y: 0, W: defaultWinSizeW, H: defaultWinSizeH}, colorB)

	// render stdin input
	for i := range renderTextSlice {
		if (*R)[i+topItem].Rank != math.MaxFloat64 {
			text, err := m.font.RenderUTF8Blended((*R)[i+topItem].Word, sdl.Color{R: rF, G: gF, B: bF, A: 255})
			if err != nil {
				return err
			}
			renderTextSlice[i] = text
		}
	}

	// draw normal item
	for i := range renderTextSlice {
		if err := renderTextSlice[i].Blit(nil, backGRender, &sdl.Rect{X: 1, Y: 1 + int32((i * fontSize)), W: 0, H: 0}); err != nil {
			return err
		}
	}

	backGRender.Blit(nil, m.surface, &sdl.Rect{X: 0, Y: 0, W: defaultWinSizeW, H: defaultWinSizeH})

	for _, sur := range renderTextSlice {
		defer sur.Free()
	}

	defer backGRender.Free()
	return
}
