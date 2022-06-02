package draw

import (
	"fmt"
	"math"
	"os"

	"github.com/Flaneur3434/go-menu/util"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	fontSize        = 16
	defaultWinSizeH = 480
	defaultWinSizeW = 640
)

type Menu struct {
	window        *sdl.Window
	renderer      *sdl.Renderer
	surface       *sdl.Surface
	font          *ttf.Font
	numOfRows     int
	ItemList      []string
	KeyBoardInput string
}

func SetUpMenu(fontPath string, menuChan chan *Menu, errChan chan error) {
	m := Menu{window: nil, surface: nil, font: nil, numOfRows: int(defaultWinSizeH / fontSize), ItemList: nil}
	var err error

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

func (m *Menu) WriteItem(R util.Ranks) error {
	renderTextSlice := make([]*sdl.Surface, len(m.ItemList))
	var numOfItemsToDraw int
	if m.numOfRows <= len(renderTextSlice) {
		numOfItemsToDraw = m.numOfRows
	} else {
		numOfItemsToDraw = len(renderTextSlice)
	}

	// clear clear of any artifacts
	m.renderer.Clear()
	m.surface.FillRect(&sdl.Rect{X: 0, Y: 0, W: defaultWinSizeW, H: defaultWinSizeH}, 0)

	// render stdin input
	for i := 0; i < numOfItemsToDraw; i++ {
		if R[i].Rank != math.MaxFloat64 {
			text, err := m.font.RenderUTF8Blended(R[i].Word, sdl.Color{R: 255, G: 0, B: 0, A: 255})
			if err != nil {
				return err
			}

			renderTextSlice[i] = text
		}
	}

	// draw stdin input
	for i := 0; i < numOfItemsToDraw; i++ {
		if err := renderTextSlice[i].Blit(nil, m.surface, &sdl.Rect{X: 1, Y: 1 + int32((i * fontSize)), W: 0, H: 0}); err != nil {
			return err
		}
	}

	for _, sur := range renderTextSlice {
		defer sur.Free()
	}

	m.window.UpdateSurface()

	return nil
}

// name ...
func (m *Menu) WriteKeyBoard() error {
	// clear clear of any artifacts
	m.renderer.Clear()
	m.surface.FillRect(&sdl.Rect{X: 1, Y: (defaultWinSizeH/fontSize)*fontSize + defaultWinSizeH/fontSize, W: defaultWinSizeW, H: fontSize}, 0)

	if len(m.KeyBoardInput) > 0 {
		// render keyboard input
		text, err := m.font.RenderUTF8Blended(m.KeyBoardInput, sdl.Color{R: 0, G: 255, B: 0, A: 255})
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
