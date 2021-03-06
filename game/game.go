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
	Player      *Player
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

	b.initPlayer(-1, -1)

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
func (b *Board) resetHands(initialBet float64) {
	if initialBet < 0 {
		initialBet = 5
	}
	if b.Dealer == nil {
		b.Dealer = &Dealer{}
	}
	b.Dealer.hand = &cards.Hand{}
	b.Player.Bets = append(
		[]*Bet{},
		&Bet{big.NewFloat(initialBet), &cards.Hand{}, false},
	)
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

// Render gets a rendering of the dealer's Hand as a string.
func (d Dealer) Render() string {
	return d.hand.Render()
}

// Deal initial cards for the dealer and the player.
func (b *Board) Deal() *Board {
	b.Stage = &Observing{}

	// Dealer's first card
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceUp()
		b.Log.Push(fmt.Sprintf("Dealer dealt %s", card.Render()))
		b.Dealer.hand.Hit(card)
		return true
	}).Wait()

	// Player first card
	b.HitPlayer().Wait()

	// Dealer second card
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceDown()
		b.Log.Push(fmt.Sprintf("Dealer dealt %s", card.Render()))
		b.Dealer.hand.Hit(card)
		return true
	}).Wait().Wait()

	// Player second card
	b.HitPlayer().Wait()

	// Begin player stage if we have not gone straight to conclusion due to
	// player immediately getting blackjack.
	if !b.Player.ActiveBet().Hand.HasBlackJack() {
		b.ChangeStage(&PlayerStage{})
	}

	return b
}

// HitDealer hits the dealer's Hand and checks if that ends their turn.
func (b *Board) HitDealer() *Board {
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceUp()
		b.Log.Push(fmt.Sprintf("Dealer dealt %s", card.Render()))
		b.Dealer.hand.Hit(card)

		return true
	}).Wait()

	return b
}

// ConcludeDealerTurn checks the dealer's Hand and advances the game stage once they
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

// HitPlayer hits the player's first Hand and advances the game stage if
// appropriate.
func (b *Board) HitPlayer() *Board {
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceUp()
		b.Log.Push(fmt.Sprintf("Player dealt %s", card.Render()))
		b.Player.ActiveBet().Hand.Hit(card)

		return true
	}).Wait()

	// Has player bust?
	if b.Player.ActiveBet().Hand.IsBust() {
		b.Log.Push(fmt.Sprintf(
			"Player busts at %d",
			util.MinInt(b.Player.ActiveBet().Hand.Scores()),
		))
	}

	// Has player got blackjack?
	if b.Player.ActiveBet().Hand.HasBlackJack() {
		b.Log.Push("[Player has blackjack!](fg-cyan)")
	}

	b.AssessPlayerStage()

	return b
}

// DoubleDown doubles the player's first bet, hits the player's first Hand and
// immediately advances the game stage.
func (b *Board) DoubleDown() *Board {

	// Double bet.
	b.action(func(b *Board) bool {
		amount := *b.Player.Bets[0].amount
		b.Player.Bets[0].amount.Add(&amount, &amount)
		b.Player.Balance.Sub(b.Player.Balance, &amount)
		return true
	}).Wait()

	// Hit player.
	b.action(func(b *Board) bool {
		card := b.Deck.Pop().FaceUp()
		b.Log.Push(fmt.Sprintf("Player dealt %s", card.Render()))
		b.Player.ActiveBet().Hand.Hit(card)

		return true
	}).Wait()

	// Has player bust?
	if b.Player.ActiveBet().Hand.IsBust() {
		b.Log.Push(fmt.Sprintf(
			"Player busts at %d",
			util.MinInt(b.Player.ActiveBet().Hand.Scores()),
		))
	}

	// Has player got blackjack?
	if b.Player.ActiveBet().Hand.HasBlackJack() {
		b.Log.Push("[Player has blackjack!](fg-cyan)")
	}

	// Always end a hand after doubling down on it.
	b.Player.ActiveBet().stand = true

	b.AssessPlayerStage()

	return b
}

// AssessPlayerStage moves to the dealer stage if the player has finished all
// of their hands.
func (b *Board) AssessPlayerStage() {
	if b.Player.IsFinished() {
		b.ChangeStage(&DealerStage{})
	}
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
// Player
//
type Player struct {
	// Indexed bets corresponding to each player Hand
	Bets    []*Bet
	Balance *big.Float
}

// initPlayer constructs a new p instance for the board.
func (b *Board) initPlayer(initialBet float64, balance float64) *Player {
	if initialBet < 0 {
		initialBet = 5
	}
	if balance < 0 {
		balance = 95
	}
	b.Player = &Player{}
	b.resetHands(initialBet)
	b.Player.Balance = big.NewFloat(balance)
	return b.Player
}

// Raise the first bet.
func (p *Player) Raise(amount float64) bool {
	if p.Balance.Cmp(big.NewFloat(amount)) == 1 {
		p.Bets[0].amount = util.AddBigFloat(
			p.Bets[0].amount,
			amount,
		)
		p.Balance = util.AddBigFloat(
			p.Balance,
			-amount,
		)
		return true
	}
	return false
}

// Lower the first bet.
func (p *Player) Lower(amount float64) bool {
	if p.Bets[0].amount.Cmp(big.NewFloat(amount)) == 1 {
		p.Bets[0].amount = util.AddBigFloat(
			p.Bets[0].amount,
			-amount,
		)
		p.Balance = util.AddBigFloat(
			p.Balance,
			amount,
		)
		return true
	}
	return false
}

// ActiveBet gets the bet currently being played.
func (p *Player) ActiveBet() *Bet {
	for _, bet := range p.Bets {
		if !bet.IsFinished() {
			return bet
		}
	}
	return p.Bets[0]
}

// IsFinished checks if the player has any hands left to play on.
func (p *Player) IsFinished() bool {
	for _, bet := range p.Bets {
		if !bet.IsFinished() {
			return false
		}
	}
	return true
}

var ac = accounting.Accounting{Symbol: "£", Precision: 2}

// Render gets a rendering of the player's hands and bets as a string.
func (p Player) Render() string {
	buffer := bytes.Buffer{}
	last := len(p.Bets) - 1
	for i, bet := range p.Bets {
		if bet.HasFocus(&p) {
			buffer.WriteString(fmt.Sprintf(
				"%s {[%s](fg-bold,fg-cyan,fg-underline)}",
				bet.Hand.Render(),
				ac.FormatMoneyBigFloat(bet.amount),
			))
		} else {
			buffer.WriteString(fmt.Sprintf(
				"[%s {%s}](fg-magenta)",
				util.StripFormatting(bet.Hand.Render()),
				ac.FormatMoneyBigFloat(bet.amount),
			))
		}
		if i != last {
			buffer.WriteString(" | ")
		}
	}
	return buffer.String()
}

//
// Bets
//
type Bet struct {
	amount *big.Float
	Hand   *cards.Hand
	stand  bool
}

// IsFinished shows if the bet is finished, i.e. its Hand is complete and the
// player can not take further action on it.
func (b *Bet) IsFinished() bool {
	if b.stand || b.Hand.HasBlackJack() || b.Hand.IsBust() {
		return true
	}
	return false
}

// HasFocus shows if the bet is the current focus, i.e. the one the player is
// currently playing on.
func (b *Bet) HasFocus(p *Player) bool {
	if b.IsFinished() {
		return false
	}
	seenThisBet := false
	for _, bet := range p.Bets {
		// If there's an active bet before this one, then this one does not have
		// focus.
		if bet == b {
			seenThisBet = true
		}
		if (!seenThisBet) && (!bet.IsFinished()) {
			return false
		}
	}
	return true
}

// Split turns this bet and Hand into two separate bets and hands.
func (b *Bet) Split(board *Board) {
	newBet := &Bet{&*b.amount, &cards.Hand{}, false}

	board.Player.Balance.Sub(board.Player.Balance, b.amount)
	board.Player.Bets = append(board.Player.Bets, newBet)

	// Split the two cards between the bets.
	newBet.Hand.Cards = []*cards.Card{b.Hand.Cards[1]}
	b.Hand.Cards = []*cards.Card{b.Hand.Cards[0]}
}

// Conclude ends the bet and pays the player's winnings (if any).
func (b *Bet) Conclude(board *Board) {
	// Pay the winnings for this bet, if any.
	amount := *b.amount
	winnings := amount.Mul(b.amount, b.Hand.WinFactor(board.Dealer.hand))
	board.Player.Balance.Add(board.Player.Balance, winnings)
	winFactor, _ := b.Hand.WinFactor(board.Dealer.hand).Float64()
	switch winFactor {
	case 2.5:
		board.Log.Push(
			fmt.Sprintf(
				"[Player wins %s with blackjack](fg-cyan)",
				ac.FormatMoneyBigFloat(winnings)),
		)
		break
	case 2:
		board.Log.Push(
			fmt.Sprintf(
				"[Player wins %s](fg-green)",
				ac.FormatMoneyBigFloat(winnings),
			),
		)
		break
	case 1:
		board.Log.Push(
			fmt.Sprintf(
				"[Player gets %s back](fg-yellow)",
				ac.FormatMoneyBigFloat(winnings),
			),
		)
		break
	case 0:
		board.Log.Push(
			fmt.Sprintf(
				"[Player loses %s](fg-red)",
				ac.FormatMoneyBigFloat(b.amount),
			),
		)
	}
	// Reset the bet balance.
	b.amount = big.NewFloat(0)
}
