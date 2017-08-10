package game

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"sync"

	"github.com/hughgrigg/blackjack/cards"
	"github.com/hughgrigg/blackjack/util"
	"github.com/leekchan/accounting"
)

// The main game controller object
type Board struct {
	Deck        *cards.Deck
	Dealer      *Dealer
	Player      *cards.HandSet
	BetsBalance *BetsBalance
	Log         *Log
	Stage       Stage
	actionQueue chan Action
	wg          sync.WaitGroup
}

type Action func(b *Board) bool
type PlayerAction struct {
	Execute     Action
	Description string
}
type ActionSet map[string]PlayerAction

func (b *Board) Begin(actionDelay int) {
	b.Stage = Betting{}
	b.Log = &Log{}
	b.BetsBalance = newBetsBalance(-1, -1)

	b.Deck = &cards.Deck{}
	b.Deck.Init()
	b.Deck.Shuffle(cards.UniqueShuffle)

	b.resetHands()

	// Run board actions with a more human interval so the player can keep up
	b.actionQueue = make(chan Action, 99)
	go func() {
		for action := range b.actionQueue {
			time.Sleep(time.Duration(actionDelay) * time.Millisecond)
			action(b)
			b.wg.Done()
		}
	}()
}

func (b *Board) action(a Action) {
	b.wg.Add(1)
	b.actionQueue <- a
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

type Dealer struct {
	hand *cards.Hand
}

func (d Dealer) Render() string {
	return d.hand.Render()
}

func (b *Board) Deal() {
	b.Stage = &Observing{}
	b.resetHands()

	// Dealer's first card
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceUp()
		b.Log.Push(fmt.Sprintf("Dealer dealt %s", card.Render()))
		b.Dealer.hand.Hit(card)
		return true
	})

	// Player first card
	b.HitPlayer()

	// Dealer second card
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceDown()
		b.Log.Push(fmt.Sprintf("Dealer dealt %s", card.Render()))
		b.Dealer.hand.Hit(card)
		return true
	})

	// Player second card
	b.HitPlayer()

	// Begin player stage
	b.action(func(b *Board) bool {
		b.Stage = &PlayerStage{}
		return true
	})
}

func (b *Board) HitPlayer() {
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceUp()
		b.Log.Push(fmt.Sprintf("Player dealt %s", card.Render()))
		b.Player.Hands[len(b.Player.Hands)-1].Hit(card)

		// Has player bust?
		if b.Player.Hands[0].IsBust() {
			b.Log.Push(fmt.Sprintf(
				"Player busts at %d",
				util.MinInt(b.Player.Hands[0].Scores()),
			))
			b.Stage = &Observing{}
			b.action(func(b *Board) bool {
				b.Stage = &DealerStage{}
				return true
			})
		}

		// Has player got blackjack?
		if b.Player.Hands[0].HasBlackJack() {
			b.Log.Push("Player has blackjack!")
			b.Stage = &Observing{}
			b.action(func(b *Board) bool {
				b.Stage = &DealerStage{}
				return true
			})
		}

		return true
	})
}

//
// Game log
//
type Log struct {
	events []string
	limit  int
}

func (l *Log) Push(event string) {
	l.events = append(l.events, event)
	if l.limit == 0 {
		l.limit = 20
	}
	if len(l.events) > l.limit {
		l.events = l.events[len(l.events)-l.limit:]
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
// Bets and balance
// todo: rename this to "bank"?
//
type BetsBalance struct {
	// Indexed bets corresponding to each player hand
	Bets    []*big.Float
	Balance *big.Float
}

func newBetsBalance(bets float64, balance float64) *BetsBalance {
	if bets < 0 {
		bets = 5
	}
	if balance < 0 {
		balance = 95
	}
	bb := BetsBalance{}
	bb.Bets = append([]*big.Float{}, big.NewFloat(bets))
	bb.Balance = big.NewFloat(balance)
	return &bb
}

func (bb *BetsBalance) Raise(amount float64) bool {
	if bb.Balance.Cmp(big.NewFloat(amount)) == 1 {
		bb.Bets[0] = util.AddBigFloat(
			bb.Bets[0],
			amount,
		)
		bb.Balance = util.AddBigFloat(
			bb.Balance,
			-amount,
		)
		return true
	}
	return false
}

func (bb *BetsBalance) Lower(amount float64) bool {
	if bb.Bets[0].Cmp(big.NewFloat(amount)) == 1 {
		bb.Bets[0] = util.AddBigFloat(
			bb.Bets[0],
			-amount,
		)
		bb.Balance = util.AddBigFloat(
			bb.Balance,
			amount,
		)
		return true
	}
	return false
}

var ac = accounting.Accounting{Symbol: "£", Precision: 2}

func (bb BetsBalance) Render() string {
	buffer := bytes.Buffer{}
	last := len(bb.Bets) - 1
	for i, bet := range bb.Bets {
		buffer.WriteString(fmt.Sprintf(
			"[%s](fg-bold,fg-cyan)", ac.FormatMoneyBigFloat(bet),
		))
		if i != last {
			buffer.WriteString(" , ")
		}
	}
	buffer.WriteString(fmt.Sprintf(" / %s", ac.FormatMoneyBigFloat(bb.Balance)))
	return buffer.String()
}
