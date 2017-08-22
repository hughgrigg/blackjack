package ui

import (
	"testing"

	"github.com/gizak/termui"
	"github.com/hughgrigg/blackjack/game"
	"github.com/stretchr/testify/assert"
)

//
// Display
//

// Should be able to initialise the display.
func TestDisplay_Init(t *testing.T) {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	// We're just exercising the display code here more than testing it.
	display := Display{}
	display.Init()
}

// Should be able to add a new view to the display.
func TestDisplay_NewView(t *testing.T) {
	display := Display{}
	firstCount := len(display.views)
	display.NewView("foobar", 5)
	assert.Equal(t, firstCount+1, len(display.views))
}

// Should be able to attach the game board to the display.
func TestDisplay_AttachBoard(t *testing.T) {
	display := Display{}
	display.initViews()

	board := &game.Board{}

	display.AttachBoard(board)

	assert.Equal(t, board.Deck, display.deckView.renderer)
}

// Should be able to render the display as a string.
func TestDisplay_Render(t *testing.T) {
	display := Display{}

	// We're just exercising the display code here more than testing it.
	display.Render()
}

//
// View
//

// An action set renderer should be able to render an action set.
func TestActionSetRenderer_Render(t *testing.T) {
	board := &game.Board{}
	board.Stage = fooStage{}

	actionSetRenderer := ActionSetRenderer{board}

	assert.Equal(
		t,
		"[f](fg-bold,fg-green): Foobar | [q](fg-bold,fg-green): Quit",
		actionSetRenderer.Render(),
	)
}

type fooStage struct {
}

func (fs fooStage) Begin(b *game.Board) {
}

func (fs fooStage) Actions() game.ActionSet {
	return game.ActionSet{
		"f": {
			Execute: func(b *game.Board) bool {
				return true
			},
			Description: "Foobar",
		},
	}
}

// A null rendered should render to an empty string.
func TestNullRenderer_Render(t *testing.T) {
	nullRenderer := NullRenderer{}
	assert.Equal(t, "", nullRenderer.Render())
}
