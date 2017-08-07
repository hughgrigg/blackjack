package ui

import (
	"fmt"

	"bytes"
	"sort"

	"github.com/gizak/termui"
	"github.com/hughgrigg/blackjack/game"
	"github.com/hughgrigg/blackjack/util"
)

type Display struct {
	board        *game.Board
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

	// q is always quit
	termui.Handle("/sys/kbd/q", func(event termui.Event) {
		termui.StopLoop()
	})

	// pass key presses to actions for the game board's current stage
	actionKeys := []string{"d"}
	for _, actionKey := range actionKeys {
		termui.Handle(
			fmt.Sprintf("/sys/kbd/%s", actionKey),
			func(termui.Event) {
				actions := d.board.Stage.Actions()
				if playerAction, ok := actions[actionKey]; ok {
					playerAction.Execute(d.board)
				}
			},
		)
	}
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
func (d *Display) AttachBoard(b *game.Board) {
	d.board = b

	d.deckView.renderer = b.Deck
	d.dealerView.renderer = b.Dealer
	d.playerView.renderer = b.Player
	d.eventLogView.renderer = b.Log
	d.actionsView.renderer = ActionSetRenderer{b}
}

type ActionSetRenderer struct {
	board *game.Board
}

func (asr ActionSetRenderer) Render() string {
	keys := []string{}
	for k := range asr.board.Stage.Actions() {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buffer := bytes.Buffer{}
	last := keys[len(keys)-1]
	for _, k := range keys {
		buffer.WriteString(fmt.Sprintf(
			"[%s](fg-bold,fg-green): %s",
			k,
			asr.board.Stage.Actions()[k].Description),
		)
		if k != last {
			buffer.WriteString(" | ")
		}
	}
	return buffer.String()
}
