package main

import (
	"github.com/gizak/termui"
	"time"
	"github.com/hughgrigg/blackjack/cards"
	"fmt"
	"bytes"
)

func main() {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	ui := Ui{}
	ui.init()

	ui.render()
	go func() {
		for range time.Tick(time.Millisecond * 100) {
			ui.render()
		}
	}()

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Loop()
}

type Ui struct {
	board *Board
	view  *termui.Par
}

func (ui *Ui) init() {
	ui.board = &Board{}
	ui.board.deck = cards.Deck{}
	ui.board.deck.Init()
	ui.board.deck.Shuffle(cards.UniqueShuffle)
	ui.view = termui.NewPar("")
	ui.view.Float = termui.AlignCenter

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(12, 0, ui.view),
		),
	)
}

func (ui *Ui) render() {
	ui.view.Width = termui.TermWidth()
	ui.view.Height = termui.TermHeight()
	ui.view.Text = ui.board.render()
	termui.Body.Align()
	termui.Render(termui.Body)
}

type Board struct {
	deck cards.Deck
	pos int
}

func (b *Board) render() string {
	var buffer bytes.Buffer
	for _, c := range b.deck.Cards {
		buffer.WriteString(fmt.Sprintf("%s\n", c.Render()))
	}
	return buffer.String()
}
