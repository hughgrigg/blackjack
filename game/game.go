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

// ActionSet is a set of player actions for a game stage.
type ActionSet map[string]PlayerAction

// Begin initialises the board and starts its action queue.
func (b *Board) Begin(actionDelay int) *Board {
	b.Stage = Betting{}
	b.Log = &Log{}

	b.Deck = &cards.Deck{}
	b.Deck.Init()
	b.Deck.Shuffle(cards.UniqueShuffle)

	b.resetHands()
	b.initBank(-1, -1)

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

	return b
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

// ChangeStage progresses the game on to a new stage.
func (b *Board) ChangeStage(stage Stage) {
	b.Stage = &Observing{}
	b.action(func(b *Board) bool {
		b.Stage = stage
		return true
	}).Wait()
	b.Stage.Begin(b)
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
	}
	for d.MustHit() {
		b.HitDealer().Wait()
	}
	b.ConcludeDealerTurn().Wait()
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
	b.ChangeStage(&PlayerStage{})
}

// HitDealer hits the dealer's hand and checks if that ends their turn.
func (b *Board) HitDealer() *Board {
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceUp()
		b.Log.Push(fmt.Sprintf("Dealer dealt %s", card.Render()))
		b.Dealer.hand.Hit(card)

		return true
	}).Wait()

	return b
}

// ConcludeDealerTurn checks the dealer's hand and advances the game stage once they
// have hard 17 or higher.
func (b *Board) ConcludeDealerTurn() *Board {
	// Has the dealer bust?
	if b.Dealer.hand.IsBust() {
		b.Log.Push(fmt.Sprintf(
			"Dealer busts at %d",
			util.MinInt(b.Dealer.hand.Scores()),
		))
	}

	// Does the dealer have blackjack?
	if b.Dealer.hand.HasBlackJack() {
		b.Log.Push("Dealer has blackjack")
	}

	// Does the dealer have hard 17 or higher?
	if b.Dealer.hand.HasHard17() {
		b.Log.Push(fmt.Sprintf(
			"Dealer has %d",
			util.MaxInt(b.Dealer.hand.Scores()),
		))
	}

	b.ChangeStage(&Assessment{})
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
		b.ChangeStage(&DealerStage{})
	}

	// Has player got blackjack?
	if b.Player.Hands[0].HasBlackJack() {
		b.Log.Push("Player has blackjack!")
		b.ChangeStage(&DealerStage{})
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
	Bets    []*Bet
	Balance *big.Float
}

// initBank constructs a new bank instance for the board.
func (b *Board) initBank(initialBet float64, balance float64) *Bank {
	if initialBet < 0 {
		initialBet = 5
	}
	if balance < 0 {
		balance = 95
	}
	b.Bank = &Bank{}
	b.Bank.Bets = append([]*Bet{}, &Bet{big.NewFloat(initialBet), b.Player.Hands[0]})
	b.Bank.Balance = big.NewFloat(balance)
	return b.Bank
}

// Raise the first bet.
func (bank *Bank) Raise(amount float64) bool {
	if bank.Balance.Cmp(big.NewFloat(amount)) == 1 {
		bank.Bets[0].amount = util.AddBigFloat(
			bank.Bets[0].amount,
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
	if bank.Bets[0].amount.Cmp(big.NewFloat(amount)) == 1 {
		bank.Bets[0].amount = util.AddBigFloat(
			bank.Bets[0].amount,
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
			"[%s](fg-bold,fg-cyan)", ac.FormatMoneyBigFloat(bet.amount),
		))
		if i != last {
			buffer.WriteString(" , ")
		}
	}
	buffer.WriteString(fmt.Sprintf(" / %s", ac.FormatMoneyBigFloat(bank.Balance)))
	return buffer.String()
}

//
// Bets
//
type Bet struct {
	amount *big.Float
	hand   *cards.Hand
}

// Conclude finishes the bet and pays the player's winnings (if any).
func (b *Bet) Conclude(board *Board) {
	// Pay the winnings for this bet, if any.
	amount := *b.amount
	winnings := amount.Mul(
		b.amount,
		b.hand.WinFactor(board.Dealer.hand),
	)
	board.Bank.Balance.Add(board.Bank.Balance, winnings)
	switch winnings.Cmp(big.NewFloat(0)) {
	case 1:
		board.Log.Push(
			fmt.Sprintf("Player takes %s", ac.FormatMoneyBigFloat(winnings)),
		)
	case 0:
		board.Log.Push(
			fmt.Sprintf("Player loses %s", ac.FormatMoneyBigFloat(b.amount)),
		)
	}
	// Reset the bet balance.
	b.amount = big.NewFloat(0)
}
