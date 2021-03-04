package main

import (
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

var live = color.White
var dead = color.Black

// GameOfLife ...
type GameOfLife struct {
	size   int
	width  int
	height int
	state  []color.Color
	next   []color.Color
	offset []int
}

// NewGameOfLife creates a new game with width and height,
// it also serves as our initialization function 'setup()'
func NewGameOfLife(width, height int) *GameOfLife {
	game := new(GameOfLife)
	game.width = width
	game.height = height
	game.size = width * height
	game.state = make([]color.Color, game.size)
	game.next = make([]color.Color, game.size)
	game.offset = []int{
		-width - 1, // nw
		-width,     // n
		-width + 1, // ne
		1,          // e
		width + 1,  // se
		width,      // s
		width - 1,  // sw
		-1,         // w
	}
	game.Seed(25)
	return game
}

// Seed fills the state with dead cells and randomly seeds
// the state with live cells up to percentage
func (game *GameOfLife) Seed(percentage float64) {
	for i := range game.state {
		game.state[i] = dead
	}
	living := int(float64(game.size) * percentage / 100)
	for i := 0; i < living; i++ {
		game.state[rand.Intn(game.size)] = live
	}
}

// Step creates the next generation of cells
func (game *GameOfLife) Step() {
	for i := range game.state {
		neighbours := 0
		for _, j := range game.offset {
			neighbours += game.At(i + j)
		}
		if game.state[i] == live && neighbours < 2 {
			game.next[i] = dead
		} else if game.state[i] == live && neighbours > 3 {
			game.next[i] = dead
		} else if game.state[i] == dead && neighbours == 3 {
			game.next[i] = live
		} else {
			game.next[i] = game.state[i]
		}
	}
	game.state, game.next = game.next, game.state
}

// At returns the cell 'status' at a given index (1D)
func (game *GameOfLife) At(i int) int {
	if i < 0 {
		i += game.size
	}
	if i >= game.size {
		i -= game.size
	}
	if game.state[i] == live {
		return 1
	}
	return 0
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (game *GameOfLife) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		game.Seed(25)
	}
	game.Step()
	return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (game *GameOfLife) Draw(screen *ebiten.Image) {
	for i := range game.state {
		screen.Set(i%game.width, i/game.width, game.state[i])
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (game *GameOfLife) Layout(width, height int) (int, int) {
	return game.width, game.height
}

func main() {
	game := NewGameOfLife(180, 120)
	ebiten.SetWindowSize(game.width, game.height)
	ebiten.SetWindowTitle("Simple Life")
	ebiten.SetMaxTPS(10)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
