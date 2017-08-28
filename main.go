package main

import (
	"time"

	"github.com/gizak/termui"
	"github.com/hughgrigg/blackjack/game"
	"github.com/hughgrigg/blackjack/ui"
)

func main() {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	board := newBoard()

	display := newDisplay()
	display.AttachBoard(board)

	termui.Loop()
}

func newBoard() *game.Board {
	board := &game.Board{}
	board.Begin(500)
	return board
}

func newDisplay() *ui.Display {
	display := &ui.Display{}
	display.Init()
	display.Render()
	go func() {
		for range time.Tick(time.Millisecond * 100) {
			display.Render()
		}
	}()
	return display
}
