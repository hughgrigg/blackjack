package game

import (
	"math/big"

	"github.com/hughgrigg/blackjack/cards"
)

//
// Game stages and actions
//

type Stage interface {
	Begin(board *Board)
	Actions(board *Board) ActionSet
}

// Betting is when the player can place their bet and then ask to deal.
type Betting struct {
}

// Begin resets the hands and bets.
func (b Betting) Begin(board *Board) {
	board.Log.Push("Round started")
	board.resetHands()
	board.Deck.Init()
	board.Deck.Shuffle(cards.UniqueShuffle)
	board.Bank.Bets = []*Bet{
		{amount: big.NewFloat(0), hand: board.Player.Hands[0]},
	}
	board.Bank.Raise(5) // Try to bet if possible.
}

// Actions during betting are dealing, raising and lowering.
func (b Betting) Actions(board *Board) ActionSet {
	return map[string]PlayerAction{
		"d": {
			func(b *Board) bool {
				b.Deal()
				return true
			},
			"Deal",
		},
		"r": {
			func(b *Board) bool {
				return b.Bank.Raise(5)
			},
			"Raise",
		},
		"l": {
			func(b *Board) bool {
				return b.Bank.Lower(5)
			},
			"Lower",
		},
	}
}

// Observing is when the player can watch events unfold until the next stage,
// i.e. actions are blocked.
type Observing struct {
}

// Begin does nothing during an observation stage.
func (o Observing) Begin(board *Board) {
}

// Actions are empty during observing.
func (o Observing) Actions(board *Board) ActionSet {
	return map[string]PlayerAction{}
}

// PlayerStage is when the player can hit or stand.
type PlayerStage struct {
}

// Begin does nothing during the player stage.
func (ps PlayerStage) Begin(board *Board) {
}

// Actions are hit or stand during the player stage.
func (ps PlayerStage) Actions(board *Board) ActionSet {
	actions := map[string]PlayerAction{
		"h": {
			func(b *Board) bool {
				b.HitPlayer()
				return true
			},
			"Hit",
		},
		"s": {
			func(b *Board) bool {
				b.Stage = &Observing{}
				b.action(func(b *Board) bool {
					b.Stage = &DealerStage{}
					b.Bank.ActiveBet().stand = true
					go func() {
						b.Dealer.Play(b)
					}()
					return true
				})
				return true
			},
			"Stand",
		},
		"d": {
			func(b *Board) bool {
				b.DoubleDown()
				return true
			},
			"Double Down",
		},
	}
	if board.Bank.ActiveBet().hand.CanSplit() {
		actions["p"] = PlayerAction{
			func(b *Board) bool {
				b.Bank.ActiveBet().Split(b)
				return true
			},
			"Split",
		}
	}
	return actions
}

// DealerStage is the dealer's turn to play.
type DealerStage struct {
	Observing
}

// Begin triggers the dealer to play in the dealer stage.
func (ds DealerStage) Begin(board *Board) {
	board.Dealer.Play(board)
}

// Assessment is when bets are won or lost.
type Assessment struct {
	Observing
}

// Begin triggers the end game reckoning to take place during assessment.
func (a Assessment) Begin(board *Board) {
	for _, bet := range board.Bank.Bets {
		board.action(func(b *Board) bool {
			bet.Conclude(b)
			return true
		}).Wait()
	}
	board.ChangeStage(&Conclusion{})
}

type Conclusion struct {
}

// Begin does nothing at the conclusion of a round.
func (c Conclusion) Begin(board *Board) {
}

// Actions at the end of a round only allow a new round to be started.
func (c Conclusion) Actions(board *Board) ActionSet {
	return map[string]PlayerAction{
		"n": {
			func(b *Board) bool {
				b.ChangeStage(&Betting{})
				return true
			},
			"New round",
		},
	}
}
