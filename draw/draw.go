package draw

import (
	"fmt"
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
	if len(*R) > 0 {
		m.drawNorm(R)
		m.drawSel(R)
	}

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

// TODO: error checking, commenting, refactoring
func (m *Menu) drawSel(R *util.Ranks) (err error) {
	var idx int
	if currentItem+topItem < len(*R) {
		idx = currentItem + topItem
	} else {
		idx = len(*R) - 1
	}
	// using variables cause its easier to read
	x1 := int32(0)
	x2 := int32(defaultWinSizeW)

	var y1 int32
	if currentItem < (len(*R)-1)-topItem {
		y1 = int32(currentItem * fontSize)
	} else {
		y1 = int32(((len(*R) - 1) - topItem) * fontSize)
	}
	y2 := y1 + fontSize

	var textRender *sdl.Surface
	var backGRender *sdl.Surface
	backGRender, err = sdl.CreateRGBSurface(0, defaultWinSizeW, fontSize, 32, 0, 0, 0, 0)

	// get color
	rB, gB, bB := util.ConvertStrToInt32(m.selBackg)
	rF, gF, bF := util.ConvertStrToInt32(m.selForeg)
	colorB := sdl.MapRGB(m.surface.Format, rB, gB, bB)

	// rendering part
	backGRender.FillRect(&sdl.Rect{X: 0, Y: 0, W: defaultWinSizeW, H: fontSize}, colorB)
	textRender, err = m.font.RenderUTF8Blended((*R)[idx].Word, sdl.Color{R: rF, G: gF, B: bF, A: 255})
	textRender.Blit(nil, backGRender, &sdl.Rect{X: 0, Y: 0, W: defaultWinSizeH, H: fontSize})
	backGRender.Blit(nil, m.surface, &sdl.Rect{X: x1, Y: y1, W: x2, H: y2})
	defer textRender.Free()
	defer backGRender.Free()

	return
}

// TODO: error checking, commenting, refactoring
func (m *Menu) drawNorm(R *util.Ranks) (err error) {
	var numToRender int
	if len(*R)-topItem < m.numOfRows {
		numToRender = len(*R) - topItem
	} else {
		numToRender = m.numOfRows
	}
	var text *sdl.Surface
	var backGRender *sdl.Surface

	backGRender, err = sdl.CreateRGBSurface(0, defaultWinSizeW, defaultWinSizeH, 32, 0, 0, 0, 0)
	if err != nil {
		return
	}

	// get color
	rB, gB, bB := util.ConvertStrToInt32(m.normBackg)
	rF, gF, bF := util.ConvertStrToInt32(m.normForeg)

	colorB := sdl.MapRGB(m.surface.Format, rB, gB, bB)

	err = backGRender.FillRect(&sdl.Rect{X: 0, Y: 0, W: defaultWinSizeW, H: defaultWinSizeH}, colorB)
	if err != nil {
		return
	}

	// draw normal item
	for i := 0; i < numToRender; i++ {
		text, err = m.font.RenderUTF8Blended((*R)[i+topItem].Word, sdl.Color{R: rF, G: gF, B: bF, A: 255})
		if err != nil {
			return
		}
		err = text.Blit(nil, backGRender, &sdl.Rect{X: 1, Y: 1 + int32((i * fontSize)), W: 0, H: 0})
		if err != nil {
			return
		}
		defer text.Free()
	}

	err = backGRender.Blit(nil, m.surface, &sdl.Rect{X: 0, Y: 0, W: defaultWinSizeW, H: defaultWinSizeH})
	if err != nil {
		return
	}

	defer backGRender.Free()
	return
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

// Move the menu down by one item
func (m *Menu) ScrollMenuDown(R *util.Ranks) {
	mu.Lock()
	defer mu.Unlock()
	if currentItem == m.numOfRows-1 && topItem < len(*R) {
		// if last item on page
		topItem += m.numOfRows
		currentItem = 0
	} else if topItem == (len(*R)/m.numOfRows)*m.numOfRows {
		// if on last page
		if currentItem < len(*R)%m.numOfRows-1 {
			currentItem++
		} else {
			// if go pass last item, go to the top
			topItem = 0
			currentItem = 0
		}
	} else {
		currentItem++
	}
}

// Move the menu Up by one item
func (m *Menu) ScrollMenuUp(R *util.Ranks) {
	mu.Lock()
	defer mu.Unlock()
	if currentItem%m.numOfRows == 0 && topItem >= m.numOfRows {
		// if first item
		topItem -= m.numOfRows
		currentItem = m.numOfRows - 1
	} else if topItem == 0 && currentItem == 0 {
		// if first page first item
		topItem = (len(*R) / m.numOfRows) * m.numOfRows
		currentItem = len(*R)%m.numOfRows - 1
	} else if currentItem > 0 {
		currentItem--
	}
}

func (m *Menu) ResetPosCounters() {
	mu.Lock()
	defer mu.Unlock()
	topItem = 0
	currentItem = 0
}
