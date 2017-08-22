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
	board.Begin(50)

	start := time.Now()
	board.action(func(b *Board) bool {
		return true
	}).Wait()

	assert.True(
		t,
		time.Since(start).Nanoseconds() > 50000000, // = 50ms
		"Board action should have been delayed by 50ms",
	)
}

// The player should be able to hit and have the game proceed from there.
func TestBoard_HitPlayer(t *testing.T) {
	board := Board{}
	board.Begin(0)

	board.HitPlayer().Wait()

	// Should have added a card to the player's hand
	assert.Equal(t, 1, len(board.Player.Hands[0].Cards))
}

// The player stage should end if the player gets blackjack.
func TestBoard_HitPlayer_BlackJack(t *testing.T) {
	board := &Board{}
	board.Begin(0)

	board.Deck.ForceNext(cards.NewCard(cards.Ace, cards.Spades))
	board.HitPlayer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Jack, cards.Diamonds))
	board.HitPlayer().Wait()

	// Dealer should play through.
	assert.Equal(t, &Assessment{}, board.Stage)
}

// The player stage should end if the player busts.
func TestBoard_HitPlayer_Bust(t *testing.T) {
	board := Board{}
	board.Begin(0)

	board.Deck.ForceNext(cards.NewCard(cards.Queen, cards.Spades))
	board.HitPlayer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Jack, cards.Diamonds))
	board.HitPlayer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Three, cards.Hearts))
	board.HitPlayer().Wait()

	// Dealer should play through.
	assert.Equal(t, &Assessment{}, board.Stage)
}

// Should be able to hit the dealer's hand.
func TestBoard_HitDealer(t *testing.T) {
	board := Board{}
	board.Begin(0)
	board.Stage = &DealerStage{}

	board.HitDealer().Wait()

	assert.Equal(t, &DealerStage{}, board.Stage)
	assert.Equal(t, 1, len(board.Dealer.hand.Cards))
}

// Should advance to assessment stage if dealer busts.
func TestBoard_HitDealerBust(t *testing.T) {
	board := Board{}
	board.Begin(0)
	board.Stage = &DealerStage{}

	board.Deck.ForceNext(cards.NewCard(cards.Queen, cards.Spades))
	board.HitDealer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Queen, cards.Diamonds))
	board.HitDealer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Queen, cards.Hearts))
	board.HitDealer().Wait()

	assert.Equal(t, &Assessment{}, board.Stage)
}

// Should advance to assessment stage if dealer has hard 17.
func TestBoard_HitDealer_Hard17(t *testing.T) {
	board := Board{}
	board.Begin(0)
	board.Stage = &DealerStage{}

	board.Deck.ForceNext(cards.NewCard(cards.Ten, cards.Spades))
	board.HitDealer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Seven, cards.Diamonds))
	board.HitDealer().Wait()

	assert.Equal(t, &Assessment{}, board.Stage)
}

// Should not advance to assessment stage if dealer has soft 17.
func TestBoard_HitDealer_Soft17(t *testing.T) {
	board := Board{}
	board.Begin(0)
	board.Stage = &DealerStage{}

	board.Deck.ForceNext(cards.NewCard(cards.Ace, cards.Spades))
	board.HitDealer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Six, cards.Diamonds))
	board.HitDealer().Wait()

	assert.Equal(t, &DealerStage{}, board.Stage)
}

// Should be able to change the game stage.
func TestBoard_ChangeStage(t *testing.T) {
	board := Board{}
	board.Begin(0)

	board.ChangeStage(&PlayerStage{})

	assert.Equal(t, &PlayerStage{}, board.Stage)
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
// Bank
//

// Should be able to render the player's bank.
func TestBank_Render(t *testing.T) {
	bank := newBank(5, 95)
	assert.Equal(t, "[£5.00](fg-bold,fg-cyan) / £95.00", bank.Render())
	bank.Bets = append(bank.Bets, &Bet{big.NewFloat(2), nil})
	assert.Equal(
		t,
		"[£5.00](fg-bold,fg-cyan) , [£2.00](fg-bold,fg-cyan) / £95.00",
		bank.Render(),
	)
}

// Should be able to raise the bet.
func TestBank_Raise(t *testing.T) {
	bank := newBank(10, 15)
	raised := bank.Raise(5)
	assert.True(t, raised, "Bet should be raised")
	assert.Equal(t, big.NewFloat(15), bank.Bets[0].amount)
	assert.Equal(t, big.NewFloat(10), bank.Balance)
}

// Should be able to lower the bet.
func TestBank_Lower(t *testing.T) {
	bank := newBank(15, 0)
	lowered := bank.Lower(5)
	assert.True(t, lowered, "Bet should be lowered")
	assert.Equal(t, big.NewFloat(10), bank.Bets[0].amount)
	assert.Equal(t, big.NewFloat(5), bank.Balance)
}

// Should not be able to raise the bet beyond the available balance.
func TestBank_RaiseMax(t *testing.T) {
	bank := newBank(0, 5)
	raised := bank.Raise(10)
	assert.False(t, raised, "Bet should not be raised")
	assert.Equal(t, big.NewFloat(0), bank.Bets[0].amount)
	assert.Equal(t, big.NewFloat(5), bank.Balance)
}

// Should not be able to lower the bet beyond the minimum.
func TestBank_LowerMin(t *testing.T) {
	bank := newBank(5, 5)
	raised := bank.Lower(10)
	assert.False(t, raised, "Bet should not be lowered")
	assert.Equal(t, big.NewFloat(5), bank.Bets[0].amount)
	assert.Equal(t, big.NewFloat(5), bank.Balance)
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

	originalBetAmount := *board.Bank.Bets[0].amount

	raise := betting.Actions()["r"]
	raise.Execute(board)
	board.wg.Wait()

	// Should raise player's bet
	assert.Equal(
		t,
		board.Bank.Bets[0].amount.Cmp(&originalBetAmount),
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

	originalBetAmount := *board.Bank.Bets[0].amount

	lower := betting.Actions()["l"]
	lower.Execute(board)
	board.wg.Wait()

	// Should lower player's bet
	assert.Equal(
		t,
		board.Bank.Bets[0].amount.Cmp(&originalBetAmount),
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

	// Should advance to dealer stage.
	assert.Equal(t, &DealerStage{}, board.Stage)
}
