package ui

import (
	"github.com/gizak/termui"
	"github.com/hughgrigg/blackjack/util"
)

type Display struct {
	dealerView   *termui.Par
	deckView     *termui.Par
	handsView    *termui.Par
	eventLogView *termui.Par
	actionsView  *termui.Par
}

func (d *Display) Init() {
	d.deckView = d.NewView("Deck", 5)
	d.dealerView = d.NewView("Dealer's Hand", 5)
	d.dealerView.BorderLabelFg = termui.ColorRed
	d.handsView = d.NewView("Player Hands", 5)
	d.handsView.BorderLabelFg = termui.ColorGreen
	d.actionsView = d.NewView("Actions", 5)

	d.eventLogView = d.NewView("Events", util.SumInts([]int{
		d.deckView.Height,
		d.dealerView.Height,
		d.handsView.Height,
		d.actionsView.Height,
	}))

	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(7, 0, d.deckView, d.dealerView, d.handsView, d.actionsView),
			termui.NewCol(5, 0, d.eventLogView),
		),
	)
}

func (d *Display) NewView(label string, height int) *termui.Par {
	view := termui.NewPar("")
	view.BorderLabel = label
	view.Height = height
	view.BorderFg = termui.ColorRGB(50, 50, 50)
	view.BorderLabelFg = termui.ColorWhite
	return view
}

func (d *Display) Render() {
	//d.view.Width = termui.TermWidth()
	//d.view.Height = termui.TermHeight()
	//d.view.Text = d.board.render()
	termui.Body.Align()
	termui.Render(termui.Body)
}
