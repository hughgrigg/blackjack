package game

import (
	"fmt"
	"testing"

	"math/big"

	"github.com/hughgrigg/blackjack/cards"
	"github.com/stretchr/testify/assert"
)

//
// Dealer
//
func TestDealer_Render(t *testing.T) {
	dealer := Dealer{&cards.Hand{}}
	dealer.hand.Hit(cards.NewCard(cards.Ace, cards.Spades))
	assert.Equal(t, "A♤  (1 / 11)", dealer.Render())
}

//
// Log
//
func TestLog_Render(t *testing.T) {
	log := Log{}
	log.Push("Foo happened")
	log.Push("Bar happened")
	assert.Equal(t, "Foo happened\n Bar happened\n", log.Render())
}

func TestLog_Limit(t *testing.T) {
	log := Log{}
	log.limit = 3
	for i := 0; i < 5; i++ {
		log.Push(fmt.Sprintf("Event %d", i))
	}
	assert.Equal(t, "Event 2\n Event 3\n Event 4\n", log.Render())
}

//
// Bets and balance
//
func TestBetsBalance_Render(t *testing.T) {
	betsBalance := newBetsBalance(5, 95)
	assert.Equal(t, "[£5.00](fg-bold,fg-cyan) / £95.00", betsBalance.Render())
	betsBalance.Bets = append(betsBalance.Bets, big.NewFloat(2))
	assert.Equal(
		t,
		"[£5.00](fg-bold,fg-cyan) , [£2.00](fg-bold,fg-cyan) / £95.00",
		betsBalance.Render(),
	)
}

//
// Game stages and actions
//
