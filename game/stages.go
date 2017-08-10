package game

//
// Game stages and actions
//

type Stage interface {
	Actions() ActionSet
}

// Player can place their bet and then ask to deal
type Betting struct {
}

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

// Player can watch events unfold until the next stage (actions are blocked)
type Observing struct {
}

func (o Observing) Actions() ActionSet {
	return map[string]PlayerAction{}
}

// Player can hit or stick
type PlayerStage struct {
}

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
					return true
				})
				return true
			},
			"Stick",
		},
	}
}

// Dealer hits until > 17
type DealerStage struct {
}

// todo
func (ds DealerStage) Actions() ActionSet {
	return map[string]PlayerAction{}
}
