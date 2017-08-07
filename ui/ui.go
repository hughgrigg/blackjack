package ui

import (
	"github.com/gizak/termui"
	"github.com/hughgrigg/blackjack/util"
)

type Display struct {
	dealerView   *View
	deckView     *View
	playerView   *View
	betView      *View
	eventLogView *View
	actionsView  *View
	views        []*View
}

func (d *Display) Init() {
	d.initViews()

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})
}

func (d *Display) initViews() {
	d.deckView = d.NewView("Deck", 5)
	d.dealerView = d.NewView("Dealer's Hand", 5)
	d.dealerView.BorderLabelFg = termui.ColorRed
	d.playerView = d.NewView("Player's Hand", 5)
	d.playerView.BorderLabelFg = termui.ColorGreen
	d.betView = d.NewView("Bet / Balance", 5)
	d.actionsView = d.NewView("Actions", 5)
	d.eventLogView = d.NewView("Game Log", util.SumInts([]int{
		d.deckView.Height,
		d.dealerView.Height,
		d.playerView.Height,
		d.betView.Height,
		d.actionsView.Height,
	}))
	termui.Body.AddRows(
		termui.NewRow(
			termui.NewCol(
				7,
				0,
				d.deckView,
				d.dealerView,
				d.playerView,
				d.betView,
				d.actionsView,
			),
			termui.NewCol(5, 0, d.eventLogView),
		),
	)
}

func (d *Display) NewView(label string, height int) *View {
	view := &View{*termui.NewPar(""), NullRenderer{}}
	view.BorderLabel = label
	view.Height = height
	view.BorderFg = termui.ColorRGB(50, 50, 50)
	view.BorderLabelFg = termui.ColorWhite
	d.views = append(d.views, view)
	return view
}

func (d *Display) Render() {
	termui.Body.Align()
	for _, view := range d.views {
		view.Text = "\n " + view.renderer.Render()
	}
	termui.Render(termui.Body)
}

type View struct {
	termui.Par
	renderer Renderer
}

type Renderer interface {
	Render() string
}

type NullRenderer struct {
}

func (n NullRenderer) Render() string {
	return ""
}

// Allow setting renderer interfaces for each part of the display
func (d *Display) SetDeck(r Renderer) {
	d.deckView.renderer = r
}
func (d *Display) SetDealer(r Renderer) {
	d.dealerView.renderer = r
}
func (d *Display) SetPlayer(r Renderer) {
	d.playerView.renderer = r
}
func (d *Display) SetBet(r Renderer) {
	d.betView.renderer = r
}
func (d *Display) SetActions(r Renderer) {
	d.actionsView.renderer = r
}
