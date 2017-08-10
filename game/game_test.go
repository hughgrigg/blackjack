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

//
// Betting
//
func TestBetting_Actions_Deal(t *testing.T) {
	betting := Betting{}

	board := &Board{}
	board.Begin(0)

	deal := betting.Actions()["d"]
	deal.Execute(board)
	board.wg.Wait()

	// Should move on to player stage after dealing
	assert.IsType(t, &PlayerStage{}, board.Stage)
}

func TestBetting_Actions_Raise(t *testing.T) {
	betting := Betting{}

	board := &Board{}
	board.Begin(0)

	originalBet := board.BetsBalance.Bets[0]

	raise := betting.Actions()["r"]
	raise.Execute(board)
	board.wg.Wait()

	// Should raise player's bet
	assert.Equal(
		t,
		board.BetsBalance.Bets[0].Cmp(originalBet),
		1,
		"Bet should have been raised",
	)
}

func TestBetting_Actions_Lower(t *testing.T) {
	betting := Betting{}

	board := &Board{}
	board.Begin(0)

	raise := betting.Actions()["r"]
	raise.Execute(board)
	originalBet := board.BetsBalance.Bets[0]

	lower := betting.Actions()["l"]
	lower.Execute(board)
	board.wg.Wait()

	// Should lower player's bet
	assert.Equal(
		t,
		board.BetsBalance.Bets[0].Cmp(originalBet),
		-1,
		"Bet should have been lowered",
	)
}

//
// Player stage
//
func TestPlayerStage_Actions_Hit(t *testing.T) {
	playerStage := PlayerStage{}

	board := &Board{}
	board.Begin(0)

	originalHandSize := len(board.Player.Hands[0].Cards)

	hit := playerStage.Actions()["h"]
	hit.Execute(board)
	board.wg.Wait()

	// Should increase player hand size
	assert.True(
		t,
		len(board.Player.Hands[0].Cards) > originalHandSize,
		"Size of player's hand should have increased",
	)
}

func TestPlayerStage_Actions_Stick(t *testing.T) {
	playerStage := PlayerStage{}

	board := &Board{}
	board.Begin(0)

	stick := playerStage.Actions()["s"]
	stick.Execute(board)
	board.wg.Wait()

	// Should move on to dealer stage
	assert.IsType(t, &DealerStage{}, board.Stage)
}
