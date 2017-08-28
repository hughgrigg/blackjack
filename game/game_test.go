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
	board.Begin(0).Wait()

	board.HitPlayer().Wait()

	// Should have added a card to the player's Hand
	assert.Equal(t, 1, len(board.Player.ActiveBet().Hand.Cards))
}

// The player stage should end if the player gets blackjack.
func TestBoard_HitPlayer_BlackJack(t *testing.T) {
	board := &Board{}
	board.Begin(0).Wait()

	board.Deck.ForceNext(cards.NewCard(cards.Ace, cards.Spades))
	board.HitPlayer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Jack, cards.Diamonds))
	board.HitPlayer().Wait()

	// Dealer should play through and the round should finish.
	assert.Equal(t, &Conclusion{}, board.Stage)
}

// The player stage should end if the player busts.
func TestBoard_HitPlayer_Bust(t *testing.T) {
	board := Board{}
	board.Begin(0).Wait()

	board.Deck.ForceNext(cards.NewCard(cards.Queen, cards.Spades))
	board.HitPlayer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Jack, cards.Diamonds))
	board.HitPlayer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Three, cards.Hearts))
	board.HitPlayer().Wait()

	// Dealer should play through and the round should finish.
	assert.Equal(t, &Conclusion{}, board.Stage)
}

// Should be able to double down.
func TestBoard_DoubleDown(t *testing.T) {
	board := Board{}
	board.Begin(0).Wait()

	// Force player win
	board.Deck.Cards[51] = cards.NewCard(cards.Ten, cards.Clubs)     // dealer 1
	board.Deck.Cards[50] = cards.NewCard(cards.Queen, cards.Spades)  // player 1
	board.Deck.Cards[49] = cards.NewCard(cards.Seven, cards.Clubs)   // dealer 2
	board.Deck.Cards[48] = cards.NewCard(cards.Five, cards.Diamonds) // player 2
	board.Deck.Cards[47] = cards.NewCard(cards.Five, cards.Hearts)   // player 3
	board.Deal().Wait()

	originalBet := *board.Player.Bets[0].amount
	originalBalance := *board.Player.Balance

	board.DoubleDown().Wait()

	// Dealer should play through and the round should finish.
	assert.Equal(t, &Conclusion{}, board.Stage)

	// The player should have won with a doubled bet.
	assert.Equal(
		t,
		originalBalance.Add(
			&originalBalance,
			originalBet.Mul(
				&originalBet,
				big.NewFloat(3), // = 2x bet, plus the bet itself
			),
		),
		board.Player.Balance,
	)
}

// Should be able to hit the dealer's Hand.
func TestBoard_HitDealer(t *testing.T) {
	board := Board{}
	board.Begin(0).Wait()
	board.Stage = &DealerStage{}

	board.HitDealer().Wait()

	assert.Equal(t, &DealerStage{}, board.Stage)
	assert.Equal(t, 1, len(board.Dealer.hand.Cards))
}

// Should advance to assessment stage if dealer busts.
func TestBoard_HitDealerBust(t *testing.T) {
	board := Board{}
	board.Begin(0).Wait()
	board.Stage = &DealerStage{}

	board.Deck.ForceNext(cards.NewCard(cards.Queen, cards.Spades))
	board.HitDealer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Queen, cards.Diamonds))
	board.HitDealer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Queen, cards.Hearts))
	board.HitDealer().Wait()

	board.ConcludeDealerTurn().Wait()

	// Dealer should play through and the round should finish.
	assert.Equal(t, &Conclusion{}, board.Stage)
}

// Should advance to assessment stage if dealer has hard 17.
func TestBoard_HitDealer_Hard17(t *testing.T) {
	board := Board{}
	board.Begin(0).Wait()
	board.Stage = &DealerStage{}

	board.Deck.ForceNext(cards.NewCard(cards.Ten, cards.Spades))
	board.HitDealer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Seven, cards.Diamonds))
	board.HitDealer().Wait()

	board.ConcludeDealerTurn().Wait()

	// Dealer should play through and the round should finish.
	assert.Equal(t, &Conclusion{}, board.Stage)
}

// Dealer must hit on soft 17.
func TestBoard_HitDealer_Soft17(t *testing.T) {
	board := Board{}
	board.Begin(0).Wait()
	board.Stage = &DealerStage{}

	board.Deck.ForceNext(cards.NewCard(cards.Ace, cards.Spades))
	board.HitDealer().Wait()
	board.Deck.ForceNext(cards.NewCard(cards.Six, cards.Diamonds))
	board.HitDealer().Wait()

	board.ConcludeDealerTurn().Wait()

	assert.True(t, board.Dealer.MustHit())
}

// Should be able to change the game stage.
func TestBoard_ChangeStage(t *testing.T) {
	board := Board{}
	board.Begin(0).Wait()

	board.ChangeStage(&PlayerStage{})

	assert.Equal(t, &PlayerStage{}, board.Stage)
}

//
// Dealer
//

// Should be able to render the dealer's Hand.
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
// Player
//

// Should be able to render the player's hands and bets.
func TestPlayer_Render(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(5, 95)
	assert.Equal(t, "(0) {[£5.00](fg-bold,fg-cyan,fg-underline)}", player.Render())
	player.Bets = append(
		player.Bets,
		&Bet{big.NewFloat(2), &cards.Hand{}, false},
	)
	assert.Equal(
		t,
		"(0) {[£5.00](fg-bold,fg-cyan,fg-underline)} | [(0) {£2.00}](fg-magenta)",
		player.Render(),
	)
}

// Should be able to raise the bet.
func TestPlayer_Raise(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(10, 15)
	raised := player.Raise(5)
	assert.True(t, raised, "Bet should be raised")
	assert.Equal(t, big.NewFloat(15), player.Bets[0].amount)
	assert.Equal(t, big.NewFloat(10), player.Balance)
}

// Should be able to lower the bet.
func TestPlayer_Lower(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(15, 0)
	lowered := player.Lower(5)
	assert.True(t, lowered, "Bet should be lowered")
	assert.Equal(t, big.NewFloat(10), player.Bets[0].amount)
	assert.Equal(t, big.NewFloat(5), player.Balance)
}

// Should not be able to raise the bet beyond the available balance.
func TestPlayer_RaiseMax(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(0, 5)
	raised := player.Raise(10)
	assert.False(t, raised, "Bet should not be raised")
	assert.Equal(t, big.NewFloat(0), player.Bets[0].amount)
	assert.Equal(t, big.NewFloat(5), player.Balance)
}

// Should not be able to lower the bet beyond the minimum.
func TestPlayer_LowerMin(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(5, 5)
	raised := player.Lower(10)
	assert.False(t, raised, "Bet should not be lowered")
	assert.Equal(t, big.NewFloat(5), player.Bets[0].amount)
	assert.Equal(t, big.NewFloat(5), player.Balance)
}

// Should be able to get the bet being played.
func TestPlayer_ActiveBet(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(5, 5)
	assert.IsType(t, &Bet{}, player.ActiveBet())
}

// The player should be seen as finished if they have no hands left to play on.
func TestPlayer_IsFinishedTrue(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(5, 5)
	player.ActiveBet().stand = true
	assert.True(t, player.IsFinished())
}

// The player should be seen as not finished if they have a hand left to play.
func TestPlayer_IsFinishedFalse(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(5, 5)
	assert.False(t, player.IsFinished())
}

//
// Bet
//

// A bet should be finished if its hand has been stood on.
func TestBet_IsFinished_Stand(t *testing.T) {
	bet := &Bet{big.NewFloat(0), &cards.Hand{}, false}
	assert.False(t, bet.IsFinished())
	bet.stand = true
	assert.True(t, bet.IsFinished())
}

// A bet should be finished if its hand has blackjack.
func TestBet_IsFinished_Blackjack(t *testing.T) {
	bet := &Bet{big.NewFloat(0), &cards.Hand{}, false}
	assert.False(t, bet.IsFinished())
	bet.Hand.Hit(cards.NewCard(cards.Ace, cards.Spades))
	bet.Hand.Hit(cards.NewCard(cards.Jack, cards.Diamonds))
	assert.True(t, bet.IsFinished())
}

// A bet should be finished if its hand is bust.
func TestBet_IsFinished_Bust(t *testing.T) {
	bet := &Bet{big.NewFloat(0), &cards.Hand{}, false}
	assert.False(t, bet.IsFinished())
	bet.Hand.Hit(cards.NewCard(cards.Queen, cards.Spades))
	bet.Hand.Hit(cards.NewCard(cards.Jack, cards.Diamonds))
	bet.Hand.Hit(cards.NewCard(cards.Three, cards.Hearts))
	assert.True(t, bet.IsFinished())
}

// A bet should have focus if it is the only bet.
func TestBet_HasFocus_Alone(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(5, 95)
	assert.True(t, player.Bets[0].HasFocus(player))
}

// A bet should have focus if it is the first bet and is not finished.
func TestBet_HasFocus_FirstNotFinished(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(5, 95)
	player.Bets = append(player.Bets, &Bet{big.NewFloat(0), &cards.Hand{}, false})
	assert.True(t, player.Bets[0].HasFocus(player))
	assert.False(t, player.Bets[1].HasFocus(player))
}

// A bet should have focus if it is second bet and the first is finished.
func TestBet_HasFocus_SecondFirstFinished(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(5, 95)
	player.Bets = append(player.Bets, &Bet{big.NewFloat(0), &cards.Hand{}, false})
	player.Bets[0].stand = true
	assert.False(t, player.Bets[0].HasFocus(player))
	assert.True(t, player.Bets[1].HasFocus(player))
}

// A bet should have focus if it is second bet and the first is finished, even
// if the third bet is not finished.
func TestBet_HasFocus_SecondFirstFinishedThirdNot(t *testing.T) {
	player := (&Board{}).Begin(0).initPlayer(5, 95)
	player.Bets = append(player.Bets, &Bet{big.NewFloat(0), &cards.Hand{}, false})
	player.Bets = append(player.Bets, &Bet{big.NewFloat(0), &cards.Hand{}, false})
	player.Bets[0].stand = true
	assert.False(t, player.Bets[0].HasFocus(player))
	assert.True(t, player.Bets[1].HasFocus(player))
	assert.False(t, player.Bets[2].HasFocus(player))
}

//
// Game stages and actions
//

//
// Betting
//

// Should be able to begin the betting stage.
func TestBetting_Begin(t *testing.T) {
	betting := Betting{}

	board := &Board{}
	board.Begin(0)

	board.ChangeStage(&betting)

	// Round should have started.
	assert.Equal(
		t,
		"Round started",
		board.Log.events[len(board.Log.events)-1],
	)
}

// The player should be able to end the betting stage by dealing.
func TestBetting_Actions_Deal(t *testing.T) {
	betting := Betting{}

	board := &Board{}
	board.Begin(0)

	// make sure we don't get blackjack
	board.Deck.ForceNext(cards.NewCard(cards.Two, cards.Diamonds))

	deal := betting.Actions(board)["d"]
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

	originalBetAmount := *board.Player.Bets[0].amount

	raise := betting.Actions(board)["r"]
	raise.Execute(board)
	board.wg.Wait()

	// Should raise player's bet
	assert.Equal(
		t,
		board.Player.Bets[0].amount.Cmp(&originalBetAmount),
		1,
		"Bet should have been raised",
	)
}

// The player should be able to lower their bet during the dealing stage.
func TestBetting_Actions_Lower(t *testing.T) {
	betting := Betting{}

	board := &Board{}
	board.Begin(0)

	raise := betting.Actions(board)["r"]
	raise.Execute(board)

	originalBetAmount := *board.Player.Bets[0].amount

	lower := betting.Actions(board)["l"]
	lower.Execute(board)
	board.wg.Wait()

	// Should lower player's bet
	assert.Equal(
		t,
		board.Player.Bets[0].amount.Cmp(&originalBetAmount),
		-1,
		"Bet should have been lowered",
	)
}

//
// Observing
//

// Starting an observing stage should not do anything.
func TestObserving_Begin(t *testing.T) {
	board := &Board{}
	observing := Observing{}
	observing.Begin(board)

	assert.Empty(t, observing.Actions(board))
}

// Player should not be able to do anything during an observing stage.
func TestObserving_Actions(t *testing.T) {
	board := &Board{}
	observing := Observing{}

	assert.Empty(t, observing.Actions(board))
}

//
// Player stage
//

// The player should be able to raise their bet during the dealing stage.
func TestPlayerStage_Actions_Hit(t *testing.T) {
	playerStage := PlayerStage{}

	board := &Board{}
	board.Begin(0)

	originalHandSize := len(board.Player.ActiveBet().Hand.Cards)

	hit := playerStage.Actions(board)["h"]
	hit.Execute(board)
	board.wg.Wait()

	// Should increase player Hand size
	assert.True(
		t,
		len(board.Player.ActiveBet().Hand.Cards) > originalHandSize,
		"Size of player's Hand should have increased",
	)
}

// The player should be able to stand during the dealing stage.
func TestPlayerStage_Actions_Stand(t *testing.T) {
	playerStage := PlayerStage{}

	board := &Board{}
	board.Begin(0)

	stand := playerStage.Actions(board)["s"]
	stand.Execute(board)
	board.wg.Wait()

	assert.Equal(t, &Observing{}, board.Stage)
}

// The player should be able to double down during the dealing stage.
func TestPlayerStage_Actions_DoubleDown(t *testing.T) {
	playerStage := PlayerStage{}

	board := &Board{}
	board.Begin(0)

	doubleDown, canDoubleDown := playerStage.Actions(board)["d"]
	assert.True(t, canDoubleDown)

	doubleDown.Execute(board)
	board.wg.Wait()

	assert.Equal(t, &Conclusion{}, board.Stage)
}

// Splitting should not be possible when the active Hand does not consist of two
// cards of the same value.
func TestPlayerStage_Actions_CanNotSplit(t *testing.T) {
	board := &Board{}
	board.Begin(0)

	// Ensure splitting is not allowed first.
	board.Deck.Cards[51] = cards.NewCard(cards.Ten, cards.Clubs)     // dealer 1
	board.Deck.Cards[50] = cards.NewCard(cards.Queen, cards.Spades)  // player 1
	board.Deck.Cards[49] = cards.NewCard(cards.Seven, cards.Clubs)   // dealer 2
	board.Deck.Cards[48] = cards.NewCard(cards.Five, cards.Diamonds) // player 2
	board.Deal().Wait()

	assert.IsType(t, &PlayerStage{}, board.Stage)

	_, canSplit := board.Stage.Actions(board)["p"]
	assert.False(t, canSplit)
}

// Splitting should be possible when the active Hand consists of two cards of
// the same value.
func TestPlayerStage_Actions_CanSplit(t *testing.T) {
	board := &Board{}
	board.Begin(0)

	// Ensure splitting is allowed.
	board.Deck.Cards[51] = cards.NewCard(cards.Ten, cards.Clubs)     // dealer 1
	board.Deck.Cards[50] = cards.NewCard(cards.Five, cards.Spades)   // player 1
	board.Deck.Cards[49] = cards.NewCard(cards.Seven, cards.Clubs)   // dealer 2
	board.Deck.Cards[48] = cards.NewCard(cards.Five, cards.Diamonds) // player 2
	board.Deal().Wait()

	assert.IsType(t, &PlayerStage{}, board.Stage)

	split, canSplit := board.Stage.Actions(board)["p"]
	assert.True(t, canSplit)

	split.Execute(board)

	assert.Len(
		t,
		board.Player.Bets,
		2,
		"Should have split into 2 bets.",
	)
}

// If the player immediately gets blackjack in their initial Hand, then the
// player stage should be skipped and we should go through to the conclusion.
func TestPlayerStage_SkippedOnBlackjack(t *testing.T) {
	board := &Board{}
	board.Begin(0)

	// Force blackjack for player.
	board.Deck.Cards[51] = cards.NewCard(cards.Ten, cards.Clubs)     // dealer 1
	board.Deck.Cards[50] = cards.NewCard(cards.Ace, cards.Spades)    // player 1
	board.Deck.Cards[49] = cards.NewCard(cards.Seven, cards.Clubs)   // dealer 2
	board.Deck.Cards[48] = cards.NewCard(cards.Jack, cards.Diamonds) // player 2

	board.Deal().Wait()

	// Dealer should play through and the round should finish.
	assert.Equal(t, &Conclusion{}, board.Stage)
}

//
// Conclusion
//

// Should be able to start a new round.
func TestConclusion_Actions_NewRound(t *testing.T) {
	board := &Board{}
	board.Begin(0)

	// Force blackjack for player.
	board.Deck.Cards[51] = cards.NewCard(cards.Ten, cards.Clubs)     // dealer 1
	board.Deck.Cards[50] = cards.NewCard(cards.Ace, cards.Spades)    // player 1
	board.Deck.Cards[49] = cards.NewCard(cards.Seven, cards.Clubs)   // dealer 2
	board.Deck.Cards[48] = cards.NewCard(cards.Jack, cards.Diamonds) // player 2

	board.Deal().Wait()

	newRound := board.Stage.Actions(board)["n"]
	newRound.Execute(board)

	// A new round should have begun.
	assert.Equal(t, &Betting{}, board.Stage)
}
