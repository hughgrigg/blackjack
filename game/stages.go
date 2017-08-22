package game

//
// Game stages and actions
//

type Stage interface {
	Begin(board *Board)
	Actions() ActionSet
}

// Betting is when the player can place their bet and then ask to deal.
type Betting struct {
}

// Begin does nothing during the betting stage.
func (b Betting) Begin(board *Board) {
}

// Actions during betting are dealing, raising and lowering.
func (b Betting) Actions() ActionSet {
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
func (o Observing) Actions() ActionSet {
	return map[string]PlayerAction{}
}

// PlayerStage is when the player can hit or stand.
type PlayerStage struct {
}

// Begin does nothing during the player stage.
func (ps PlayerStage) Begin(board *Board) {
}

// Actions are hit or stand during the player stage.
func (ps PlayerStage) Actions() ActionSet {
	return map[string]PlayerAction{
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
					go func() {
						b.Dealer.Play(b)
					}()
					return true
				})
				return true
			},
			"Stand",
		},
	}
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

}
