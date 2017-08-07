package game

import (
	"github.com/hughgrigg/blackjack/cards"
	"time"
)

// The main game controller object
type Board struct {
	Deck        *cards.Deck
	Dealer      *Dealer
	Player      *cards.HandSet
	actionQueue chan Action
}

type Action func(b *Board)

func (b *Board) Begin(actionDelay int) {
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

	b.Deal()
}

type Dealer struct {
	hand *cards.Hand
}

func (d Dealer) Render() string {
	return d.hand.Render()
}

func (b *Board) resetHands() {
	b.Dealer = &Dealer{&cards.Hand{}}
	b.Player = &cards.HandSet{Hands: []*cards.Hand{{}}}
}

func (b *Board) action(a Action) {
	b.actionQueue <- a
}

func (b *Board) Deal() {
	b.resetHands()

	// Dealer's first card
	b.action(func(b *Board) {
		card := b.Deck.Pop().FaceUp()
		b.Dealer.hand.Hit(card)
	})

	// Player first card
	for _, hand := range b.Player.Hands {
		b.action(func(b *Board) {
			card := b.Deck.Pop().FaceUp()
			hand.Hit(card)
		})
	}

	// Dealer second card
	b.action(func(b *Board) {
		card := b.Deck.Pop().FaceDown()
		b.Dealer.hand.Hit(card)
	})

	// Player second card
	for _, hand := range b.Player.Hands {
		b.action(func(b *Board) {
			card := b.Deck.Pop().FaceUp()
			hand.Hit(card)
		})
	}
}
