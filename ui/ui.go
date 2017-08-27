package ui

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/gizak/termui"
	"github.com/hughgrigg/blackjack/game"
	"github.com/hughgrigg/blackjack/util"
)

// The master display object containing all the sub-views.
type Display struct {
	board        *game.Board
	dealerView   *View
	deckView     *View
	playerView   *View
	bankView     *View
	eventLogView *View
	actionsView  *View
	views        []*View
}

// Initialise the display with its views and keyboard handlers.
func (d *Display) Init() {
	d.initViews()

	// q is always quit.
	termui.Handle("/sys/kbd/q", func(event termui.Event) {
		termui.StopLoop()
	})

	// Pass key presses to actions for the game board's current stage.
	termui.Handle(
		"/sys/kbd",
		func(e termui.Event) {
			evtKbd, ok := e.Data.(termui.EvtKbd)
			if !ok {
				return
			}
			actions := d.board.Stage.Actions(d.board)
			playerAction, ok := actions[evtKbd.KeyStr]
			if !ok {
				return
			}
			d.board.Log.Push(fmt.Sprintf(
				">> [%s](fg-bold,fg-green)",
				playerAction.Description,
			))
			playerAction.Execute(d.board)
		},
	)
}

// Initialise the view sections of the display.
func (d *Display) initViews() {
	d.deckView = d.NewView("Deck", 5)
	d.dealerView = d.NewView("Dealer's Hand", 5)
	d.dealerView.BorderLabelFg = termui.ColorRed
	d.playerView = d.NewView("Player's Hand", 5)
	d.playerView.BorderLabelFg = termui.ColorGreen
	d.bankView = d.NewView("Bets / Balance", 5)
	d.actionsView = d.NewView("Actions", 5)
	d.eventLogView = d.NewView("Game Log", util.SumInts([]int{
		d.deckView.Height,
		d.dealerView.Height,
		d.playerView.Height,
		d.bankView.Height,
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
				d.bankView,
				d.actionsView,
			),
			termui.NewCol(5, 0, d.eventLogView),
		),
	)
}

// Construct a new view in the display.
func (d *Display) NewView(label string, height int) *View {
	view := &View{*termui.NewPar(""), NullRenderer{}}
	view.BorderLabel = label
	view.Height = height
	view.BorderFg = termui.ColorRGB(50, 50, 50)
	view.BorderLabelFg = termui.ColorWhite
	d.views = append(d.views, view)
	return view
}

// Have the display render itself through termui.
func (d *Display) Render() {
	termui.Body.Align()
	for _, view := range d.views {
		view.Text = "\n " + view.renderer.Render()
	}
	termui.Render(termui.Body)
}

// Allow setting renderer interfaces for each part of the display.
func (d *Display) AttachBoard(b *game.Board) {
	d.board = b

	d.deckView.renderer = b.Deck
	d.dealerView.renderer = b.Dealer
	d.playerView.renderer = b.Player
	d.bankView.renderer = b.Bank
	d.eventLogView.renderer = b.Log
	d.actionsView.renderer = ActionSetRenderer{b}
}

//
// View
//

// A viewable section of the display.
type View struct {
	termui.Par
	renderer Renderer
}

// Something that can render itself as a string.
type Renderer interface {
	Render() string
}

// An empty renderer.
type NullRenderer struct {
}

// Get an empty string as a rendering.
func (n NullRenderer) Render() string {
	return ""
}

// A renderer for a set of player actions.
type ActionSetRenderer struct {
	board *game.Board
}

// Get a rendering of a set of player actions as a string.
func (asr ActionSetRenderer) Render() string {
	keys := []string{}
	actions := asr.board.Stage.Actions(asr.board)
	for k := range actions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// Add quit at the end of the actions
	actions["q"] = game.PlayerAction{
		Execute: func(*game.Board) bool {
			termui.StopLoop()
			return true
		},
		Description: "Quit",
	}
	keys = append(keys, "q")
	buffer := bytes.Buffer{}
	last := keys[len(keys)-1]
	for _, k := range keys {
		buffer.WriteString(fmt.Sprintf(
			"[%s](fg-bold,fg-green): %s",
			k,
			actions[k].Description,
		),
		)
		if k != last {
			buffer.WriteString(" | ")
		}
	}
	return buffer.String()
}
