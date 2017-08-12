package game

import (
	"fmt"
	"testing"

	"math/big"

	"time"

	"github.com/hughgrigg/blackjack/cards"
	"github.com/stretchr/testify/assert"
)

//
// Board
//

// Board actions should happen on at intervals so that human players can keep up
// with what's happening.
func TestBoard_ActionDelay(t *testing.T) {
	board := Board{}
	board.Begin(100)

	start := time.Now()
	board.action(func(b *Board) bool {
		return true
	})
	board.wg.Wait()

	// Action should have been delayed by 100ms
	assert.True(
		t,
		time.Since(start).Nanoseconds() > 100000000, // = 100ms
		"Board action should have been delayed by 100ms",
	)
}

// The player should be able to hit and have the game proceed from there.
func TestBoard_HitPlayer(t *testing.T) {
	board := Board{}
	board.Begin(0)

	board.HitPlayer()
	board.wg.Wait()

	// Should have added a card to the player's hand
	assert.Equal(t, 1, len(board.Player.Hands[0].Cards))
}

// The player stage should end if the player gets blackjack.
func TestBoard_HitPlayer_BlackJack(t *testing.T) {
	board := Board{}
	board.Begin(0)

	board.Deck.ForceNext(cards.NewCard(cards.Ace, cards.Spades))
	board.HitPlayer()
	board.Deck.ForceNext(cards.NewCard(cards.Jack, cards.Diamonds))
	board.HitPlayer()
	board.wg.Wait()

	assert.Equal(t, &DealerStage{}, board.Stage)
}

// The player stage should end if the player busts.
func TestBoard_HitPlayer_Bust(t *testing.T) {
	board := Board{}
	board.Begin(0)

	board.Deck.ForceNext(cards.NewCard(cards.Queen, cards.Spades))
	board.HitPlayer()
	board.Deck.ForceNext(cards.NewCard(cards.Jack, cards.Diamonds))
	board.HitPlayer()
	board.Deck.ForceNext(cards.NewCard(cards.Three, cards.Hearts))
	board.HitPlayer()
	board.wg.Wait()

	assert.Equal(t, &DealerStage{}, board.Stage)
}

//
// Dealer
//

// Should be able to render the dealer's hand.
func TestDealer_Render(t *testing.T) {
	dealer := Dealer{&cards.Hand{}}
	dealer.hand.Hit(cards.NewCard(cards.Ace, cards.Spades))
	assert.Equal(t, "A♤  (1 / 11)", dealer.Render())
}

//
// Log
//

// Should be able to render the game log.
func TestLog_Render(t *testing.T) {
	log := Log{}
	log.Push("Foo happened")
	log.Push("Bar happened")
	assert.Equal(t, "Foo happened\n Bar happened\n", log.Render())
}

// The game log should be limited to a set number of lines.
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

// Should be able to render the player's bets and balance.
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

// Should be able to raise the bet.
func TestBetsBalance_Raise(t *testing.T) {
	betsBalance := newBetsBalance(10, 15)
	raised := betsBalance.Raise(5)
	assert.True(t, raised, "Bet should be raised")
	assert.Equal(t, big.NewFloat(15), betsBalance.Bets[0])
	assert.Equal(t, big.NewFloat(10), betsBalance.Balance)
}

// Should be able to lower the bet.
func TestBetsBalance_Lower(t *testing.T) {
	betsBalance := newBetsBalance(15, 0)
	lowered := betsBalance.Lower(5)
	assert.True(t, lowered, "Bet should be lowered")
	assert.Equal(t, big.NewFloat(10), betsBalance.Bets[0])
	assert.Equal(t, big.NewFloat(5), betsBalance.Balance)
}

// Should not be able to raise the bet beyond the available balance.
func TestBetsBalance_RaiseMax(t *testing.T) {
	betsBalance := newBetsBalance(0, 5)
	raised := betsBalance.Raise(10)
	assert.False(t, raised, "Bet should not be raised")
	assert.Equal(t, big.NewFloat(0), betsBalance.Bets[0])
	assert.Equal(t, big.NewFloat(5), betsBalance.Balance)
}

// Should not be able to lower the bet beyond the minimum.
func TestBetsBalance_LowerMin(t *testing.T) {
	betsBalance := newBetsBalance(5, 5)
	raised := betsBalance.Lower(10)
	assert.False(t, raised, "Bet should not be lowered")
	assert.Equal(t, big.NewFloat(5), betsBalance.Bets[0])
	assert.Equal(t, big.NewFloat(5), betsBalance.Balance)
}

//
// Game stages and actions
//

//
// Betting
//

// The player should be able to end the betting stage by dealing.
func TestBetting_Actions_Deal(t *testing.T) {
	betting := Betting{}

	board := &Board{}
	board.Begin(0)

	// make sure we don't get blackjack
	board.Deck.ForceNext(cards.NewCard(cards.Two, cards.Diamonds))

	deal := betting.Actions()["d"]
	deal.Execute(board)
	board.wg.Wait()

	// Should move on to player stage after dealing
	assert.IsType(t, &PlayerStage{}, board.Stage)
}

// The player should be able to raise their bet during the dealing stage.
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

// The player should be able to lower their bet during the dealing stage.
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
// Observing
//

// Player should not be able to do anything during an observing stage.
func TestObserving_Actions(t *testing.T) {
	observing := Observing{}

	assert.Empty(t, observing.Actions())
}

//
// Player stage
//

// The player should be able to raise their bet during the dealing stage.
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

// The player should be able to stand during the dealing stage.
func TestPlayerStage_Actions_Stand(t *testing.T) {
	playerStage := PlayerStage{}

	board := &Board{}
	board.Begin(0)

	stand := playerStage.Actions()["s"]
	stand.Execute(board)
	board.wg.Wait()

	// Should move on to dealer stage
	assert.IsType(t, &DealerStage{}, board.Stage)
}
