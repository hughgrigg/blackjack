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

// The main game controller object.
type Board struct {
	Deck        *cards.Deck
	Dealer      *Dealer
	Player      *cards.HandSet
	Bank        *Bank
	Log         *Log
	Stage       Stage
	actionQueue chan Action
	wg          *sync.WaitGroup
}

// An action that can be made on the board, returning boolean success.
type Action func(b *Board) bool

// An action the player can consider taking, with a description.
type PlayerAction struct {
	Execute     Action
	Description string
}

// A set of player actions for a game stage.
type ActionSet map[string]PlayerAction

// Initialise the board and start its action queue.
func (b *Board) Begin(actionDelay int) {
	b.Stage = Betting{}
	b.Log = &Log{}
	b.Bank = newBank(-1, -1)

	b.Deck = &cards.Deck{}
	b.Deck.Init()
	b.Deck.Shuffle(cards.UniqueShuffle)

	b.resetHands()

	// Run board actions with a more human interval so the player can keep up.
	b.wg = &sync.WaitGroup{}
	b.actionQueue = make(chan Action, 999)
	go func() {
		for action := range b.actionQueue {
			time.Sleep(time.Duration(actionDelay) * time.Millisecond)
			action(b)
			b.wg.Done()
		}
	}()
}

// Queue up an action for the board to carry out.
func (b *Board) action(a Action) *Board {
	b.wg.Add(1)
	b.actionQueue <- a
	return b
}

func (b *Board) Wait() *Board {
	b.wg.Wait()
	return b
}

// Initialise the dealer's and player's hands.
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

// Dealer is the dealer in the blackjack game.
type Dealer struct {
	hand *cards.Hand
}

// Play has the dealer carry out their turn, hitting until hard 17 or more.
func (d *Dealer) Play(b *Board) {
	for _, card := range d.hand.Cards {
		if !card.IsFaceUp() {
			b.action(func(b *Board) bool {
				card.FaceUp()
				b.Log.Push(fmt.Sprintf("Dealer had %s", card.Render()))
				return true
			}).Wait()
		}
		b.CheckDealerTurn()
	}
	for d.MustHit() {
		b.HitDealer().Wait()
	}
}

// MustHit sees if the dealer has to keep hitting. In blackjack, the dealer
// must keeping hitting until: they have 17 or more (not including soft 17), or
// they bust.
func (d *Dealer) MustHit() bool {
	if d.hand.HasHard17() {
		return false
	}
	if d.hand.IsBust() {
		return false
	}
	return true
}

// Render gets a rendering of the dealer's hand as a string.
func (d Dealer) Render() string {
	return d.hand.Render()
}

// Deal initial cards for the dealer and the player.
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

// HitDealer hits the dealer's hand and checks if that ends their turn.
func (b *Board) HitDealer() *Board {
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceUp()
		b.Log.Push(fmt.Sprintf("Dealer dealt %s", card.Render()))
		b.Dealer.hand.Hit(card)

		return true
	}).Wait()

	return b.CheckDealerTurn()
}

// CheckDealerTurn checks the dealer's hand and advances the game stage once they
// have hard 17 or higher.
func (b *Board) CheckDealerTurn() *Board {
	// Has the dealer bust?
	if b.Dealer.hand.IsBust() {
		b.Log.Push(fmt.Sprintf(
			"Dealer busts at %d",
			util.MinInt(b.Dealer.hand.Scores()),
		))
		b.Stage = &Assessment{} // todo
		return b
	}

	// Does the dealer have blackjack?
	if b.Dealer.hand.HasBlackJack() {
		b.Log.Push("Dealer has blackjack")
		b.Stage = &Assessment{} // todo
		return b
	}

	// Does the dealer have hard 17 or higher?
	if b.Dealer.hand.HasHard17() {
		b.Log.Push(fmt.Sprintf(
			"Dealer has %d",
			util.MaxInt(b.Dealer.hand.Scores()),
		))
		b.Stage = &Assessment{} // todo
		return b
	}

	return b
}

// HitPlayer hits the player's first hand and advances the game stage if
// appropriate.
func (b *Board) HitPlayer() *Board {
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceUp()
		b.Log.Push(fmt.Sprintf("Player dealt %s", card.Render()))
		b.Player.Hands[len(b.Player.Hands)-1].Hit(card)

		return true
	}).Wait()

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
		}).Wait()

		b.Dealer.Play(b)
	}

	// Has player got blackjack?
	if b.Player.Hands[0].HasBlackJack() {
		b.Log.Push("Player has blackjack!")
		b.Stage = &Observing{}
		b.action(func(b *Board) bool {
			b.Stage = &DealerStage{}
			return true
		}).Wait()

		b.Dealer.Play(b)
	}

	return b
}

//
// Game log
//
type Log struct {
	events []string
	limit  int
}

// Add a new event to the game log.
func (l *Log) Push(event string) {
	l.events = append(l.events, event)
	if l.limit == 0 {
		l.limit = 20
	}
	if len(l.events) > l.limit {
		l.events = l.events[len(l.events)-l.limit:]
	}
}

// Get a rendering of the game log as a string.
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
// Bank
//
type Bank struct {
	// Indexed bets corresponding to each player hand
	Bets    []*big.Float
	Balance *big.Float
}

// Construct a new bank instance.
func newBank(initialBet float64, balance float64) *Bank {
	if initialBet < 0 {
		initialBet = 5
	}
	if balance < 0 {
		balance = 95
	}
	bank := Bank{}
	bank.Bets = append([]*big.Float{}, big.NewFloat(initialBet))
	bank.Balance = big.NewFloat(balance)
	return &bank
}

// Raise the first bet.
func (bank *Bank) Raise(amount float64) bool {
	if bank.Balance.Cmp(big.NewFloat(amount)) == 1 {
		bank.Bets[0] = util.AddBigFloat(
			bank.Bets[0],
			amount,
		)
		bank.Balance = util.AddBigFloat(
			bank.Balance,
			-amount,
		)
		return true
	}
	return false
}

// Lower the first bet.
func (bank *Bank) Lower(amount float64) bool {
	if bank.Bets[0].Cmp(big.NewFloat(amount)) == 1 {
		bank.Bets[0] = util.AddBigFloat(
			bank.Bets[0],
			-amount,
		)
		bank.Balance = util.AddBigFloat(
			bank.Balance,
			amount,
		)
		return true
	}
	return false
}

var ac = accounting.Accounting{Symbol: "£", Precision: 2}

// Get a rendering of the bank as a string.
func (bank Bank) Render() string {
	buffer := bytes.Buffer{}
	last := len(bank.Bets) - 1
	for i, bet := range bank.Bets {
		buffer.WriteString(fmt.Sprintf(
			"[%s](fg-bold,fg-cyan)", ac.FormatMoneyBigFloat(bet),
		))
		if i != last {
			buffer.WriteString(" , ")
		}
	}
	buffer.WriteString(fmt.Sprintf(" / %s", ac.FormatMoneyBigFloat(bank.Balance)))
	return buffer.String()
}
