package game

//
// Game stages and actions
//

type Stage interface {
	Actions() ActionSet
}

// Betting is when the player can place their bet and then ask to deal.
type Betting struct {
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
				return b.BetsBalance.Raise(5)
			},
			"Raise",
		},
		"l": {
			func(b *Board) bool {
				return b.BetsBalance.Lower(5)
			},
			"Lower",
		},
	}
}

// Observing is when the player can watch events unfold until the next stage,
// i.e. actions are blocked.
type Observing struct {
}

// Actions are empty during observing.
func (o Observing) Actions() ActionSet {
	return map[string]PlayerAction{}
}

// PlayerStage is when the player can hit or stand.
type PlayerStage struct {
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

// Assessment is when bets are won or lost.
type Assessment struct {
	Observing
}
