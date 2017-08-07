package main

import (
	"github.com/gizak/termui"
	"github.com/hughgrigg/blackjack/game"
	"github.com/hughgrigg/blackjack/ui"
	"time"
)

func main() {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	board := newBoard()
	display := newDisplay()
	linkBoardWithDisplay(board, display)

	termui.Loop()
}

func newBoard() *game.Board {
	board := &game.Board{}
	board.Begin(700)
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

func linkBoardWithDisplay(b *game.Board, d *ui.Display) {
	d.SetDeck(b.Deck)
	d.SetDealer(b.Dealer)
	d.SetPlayer(b.Player)
}
