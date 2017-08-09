package game

//
// Game stages and actions
//

type Stage interface {
	Actions() ActionSet
}

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
