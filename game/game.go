package game

import (
	"time"

	"fmt"

	"bytes"

	"github.com/hughgrigg/blackjack/cards"
)

// The main game controller object
type Board struct {
	Deck        *cards.Deck
	Dealer      *Dealer
	Player      *cards.HandSet
	Log         *Log
	Stage       Stage
	actionQueue chan Action
}

type Action func(b *Board)

func (b *Board) Begin(actionDelay int) {
	b.Stage = Betting{}
	b.Log = &Log{}

	b.Deck = &cards.Deck{}
	b.Deck.Init()
	b.Deck.Shuffle(cards.UniqueShuffle)

	b.resetHands()

	// Run board actions with a more human interval so the player can keep up
	b.actionQueue = make(chan Action, 99)
	go func() {
		for action := range b.actionQueue {
			action(b)
			time.Sleep(time.Duration(actionDelay) * time.Millisecond)
		}
	}()
}

type Dealer struct {
	hand *cards.Hand
}

func (d Dealer) Render() string {
	return d.hand.Render()
}

func (b *Board) resetHands() {
	if b.Dealer == nil {
		b.Dealer = &Dealer{}
	}
	if b.Player == nil {
		b.Player = &cards.HandSet{}
	}
	b.Dealer.hand = &cards.Hand{}
	b.Player.Hands = []*cards.Hand{{}}
}

func (b *Board) action(a Action) {
	b.actionQueue <- a
}

func (b *Board) Deal() {
	b.resetHands()

	// Dealer's first card
	b.action(func(b *Board) {
		card := b.Deck.Pop().FaceUp()
		b.Log.push(fmt.Sprintf("Dealer dealt %s", card.Render()))
		b.Dealer.hand.Hit(card)
	})

	// Player first card
	for _, hand := range b.Player.Hands {
		b.action(func(b *Board) {
			card := b.Deck.Pop().FaceUp()
			b.Log.push(fmt.Sprintf("Player dealt %s", card.Render()))
			hand.Hit(card)
		})
	}

	// Dealer second card
	b.action(func(b *Board) {
		card := b.Deck.Pop().FaceDown()
		b.Log.push(fmt.Sprintf("Dealer dealt %s", card.Render()))
		b.Dealer.hand.Hit(card)
	})

	// Player second card
	for _, hand := range b.Player.Hands {
		b.action(func(b *Board) {
			card := b.Deck.Pop().FaceUp()
			b.Log.push(fmt.Sprintf("Player dealt %s", card.Render()))
			hand.Hit(card)
		})
	}
}

//
// Game log
//
type Log struct {
	events []string
}

func (l *Log) push(event string) {
	l.events = append(l.events, event)
	if len(l.events) > 21 {
		l.events = l.events[len(l.events)-21:]
	}
}

func (l Log) Render() string {
	buffer := bytes.Buffer{}
	if len(l.events) > 0 {
		buffer.WriteString(fmt.Sprintf("%s\n", l.events[0]))
	}
	if len(l.events) > 1 {
		for _, event := range l.events[1:] {
			buffer.WriteString(fmt.Sprintf(" %s\n", event))
		}
	}
	return buffer.String()
}

//
// Game stages and actions
//
type PlayerAction struct {
	Execute     Action
	Description string
}
type ActionSet map[string]PlayerAction

type Stage interface {
	Actions() ActionSet
}

type Betting struct {
}

func (b Betting) Actions() ActionSet {
	return map[string]PlayerAction{
		"d": {
			func(b *Board) {
				b.Deal()
			},
			"Deal",
		},
	}
}
