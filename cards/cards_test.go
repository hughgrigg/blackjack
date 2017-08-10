package cards

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//
// Card
//

// Cards should have slices of values according to blackjack rules.
func TestCard_Values(t *testing.T) {
	expected := map[Rank][]int{
		Ace:   {1, 11},
		Two:   {2},
		Three: {3},
		Four:  {4},
		Five:  {5},
		Six:   {6},
		Seven: {7},
		Eight: {8},
		Nine:  {9},
		Ten:   {10},
		Jack:  {10},
		Queen: {10},
		King:  {10},
	}
	for rank, values := range expected {
		card := NewCard(rank, Spades)
		assert.Equal(t, values, card.Values())
	}
}

// Cards should be able to give a readable notation as a string.
func TestCard_Notation(t *testing.T) {
	expected := map[*Card]string{
		NewCard(Ace, Spades):     "A♤",
		NewCard(Queen, Hearts):   "Q♥",
		NewCard(Two, Clubs):      "2♧",
		NewCard(Eight, Diamonds): "8♦",
	}
	for card, drawn := range expected {
		assert.Equal(t, drawn, card.Notation())
	}
}

// A card's rank and suit should be visible when it is face up.
func TestCard_FaceUp(t *testing.T) {
	card := NewCard(Ace, Spades)
	card.FaceUp()
	assert.Equal(t, "A♤", card.Notation())
}

// A card's rank and suit should be hidden when it is face down.
func TestCard_FaceDown(t *testing.T) {
	card := NewCard(Ace, Spades)
	card.FaceDown()
	assert.Equal(t, "🂠 ?", card.Notation())
}

// Cards should be able to give an output rendering with colours.
func TestCard_Render(t *testing.T) {
	expected := map[*Card]string{
		NewCard(Ten, Clubs):     "X♧",
		NewCard(King, Spades):   "K♤",
		NewCard(Five, Hearts):   "[5♥](fg-red)",
		NewCard(Jack, Diamonds): "[J♦](fg-red)",
	}
	for card, drawn := range expected {
		assert.Equal(t, drawn, card.Render())
	}
}

//
// Deck
//

// Initialising a deck should add all the cards in order.
func TestDeck_Init(t *testing.T) {
	deck := Deck{}
	deck.Init()
	assert.Equal(t, "A♧", deck.Cards[00].Notation())
	assert.Equal(t, "A♦", deck.Cards[13].Notation())
	assert.Equal(t, "A♥", deck.Cards[26].Notation())
	assert.Equal(t, "A♤", deck.Cards[39].Notation())
}

// Should be able to get an output rendering of a deck.
func TestDeck_Render(t *testing.T) {
	deck := Deck{}
	deck.Init()
	assert.Equal(t, "🂠  ×52", deck.Render())
}

// Should be able to shuffle a deck with a specific seed.
func TestDeck_Shuffle(t *testing.T) {
	deck := Deck{}
	deck.Init()
	deck.Shuffle(42)
	assert.Equal(t, "8♥", deck.Cards[0].Notation())
	assert.Equal(t, "9♦", deck.Cards[1].Notation())
	assert.Equal(t, "8♦", deck.Cards[2].Notation())
}

// Should be able to pop the top card off the deck.
func TestDeck_Pop(t *testing.T) {
	deck := Deck{}
	deck.Init()
	deck.Shuffle(42)
	assert.Equal(t, "5♤", deck.Pop().Notation())
	assert.Equal(t, "Q♤", deck.Pop().Notation())
}

// Forcing the next card to be a card currently in the deck should bring it to
// the top.
func TestDeck_ForceNext(t *testing.T) {
	deck := Deck{}
	deck.Init()
	deck.Shuffle(UniqueShuffle)

	deck.ForceNext(NewCard(Ace, Spades))

	assert.Equal(t, NewCard(Ace, Spades), deck.Pop())
}

// Forcing the next card to be a card not currently in the deck should add it.
func TestDeck_ForceNextNew(t *testing.T) {
	deck := Deck{}
	deck.Init()
	deck.Shuffle(UniqueShuffle)

	popped := deck.Pop()
	deck.ForceNext(popped)

	assert.Equal(t, popped, deck.Pop())
}

//
// Hand
//

// Hitting a hand with a card should add that card to the hand.
func TestHand_Hit(t *testing.T) {
	hand := Hand{}
	assert.Empty(t, hand.Cards)

	hand.Hit(NewCard(Ace, Spades))
	hand.Hit(NewCard(Jack, Diamonds))
	assert.Len(t, hand.Cards, 2)
	assert.Equal(t, NewCard(Ace, Spades), hand.Cards[0])
	assert.Equal(t, NewCard(Jack, Diamonds), hand.Cards[1])
}

// Should be able to render a hand with its cards and scores.
func TestHand_Render(t *testing.T) {
	hand := Hand{}
	hand.Hit(NewCard(Ace, Spades))
	hand.Hit(NewCard(Three, Clubs))
	assert.Equal(t, "A♤, 3♧  (4 / 14)", hand.Render())
}

// A hand with blackjack should only display a score of 21.
func TestHand_RenderBlackjack(t *testing.T) {
	hand := Hand{}
	hand.Hit(NewCard(Ace, Spades))
	hand.Hit(NewCard(Jack, Clubs))
	assert.Equal(t, "A♤, J♧  (21)", hand.Render())
}

// A bust hand should only displaying the lowest bust value.
func TestHand_IsBust(t *testing.T) {
	hand := Hand{}

	hand.Hit(NewCard(Jack, Diamonds))
	hand.Hit(NewCard(Eight, Hearts))
	assert.False(t, hand.IsBust())

	hand.Hit(NewCard(Five, Clubs))
	assert.True(t, hand.IsBust())
}

// A hand should know if it has blackjack.
func TestHand_HasBlackJack(t *testing.T) {
	hand := Hand{}

	hand.Hit(NewCard(Ace, Spades))
	assert.False(t, hand.HasBlackJack())

	hand.Hit(NewCard(Jack, Diamonds))
	assert.True(t, hand.HasBlackJack())
}

// Should be able to calculate the possible scores for a hand, allowing for hard
// and soft totals (due to ace being 1 or 11).
func TestHand_Scores(t *testing.T) {
	var hand Hand

	hand = Hand{}
	assert.Equal(t, []int{0}, hand.Scores())

	hand = Hand{[]*Card{
		NewCard(Five, Hearts),
	}}
	assert.Equal(t, []int{5}, hand.Scores())

	hand = Hand{[]*Card{
		NewCard(Five, Hearts),
		NewCard(Three, Hearts),
	}}
	assert.Equal(t, []int{8}, hand.Scores())

	hand = Hand{[]*Card{
		NewCard(Jack, Hearts),
		NewCard(Queen, Hearts),
		NewCard(King, Hearts),
	}}
	assert.Equal(t, []int{30}, hand.Scores())

	hand = Hand{[]*Card{
		NewCard(Five, Hearts),
		NewCard(Ace, Hearts),
	}}
	assert.Equal(t, []int{6, 16}, hand.Scores())

	hand = Hand{[]*Card{
		NewCard(Five, Hearts),
		NewCard(Ace, Hearts),
		NewCard(Three, Hearts),
	}}
	assert.Equal(t, []int{9, 19}, hand.Scores())

	hand = Hand{[]*Card{
		NewCard(Ace, Hearts),
	}}
	assert.Equal(t, []int{1, 11}, hand.Scores())

	// Show only blackjack if there is a blackjack score
	hand = Hand{[]*Card{
		NewCard(Ace, Hearts),
		NewCard(Queen, Spades),
	}}
	assert.Equal(t, []int{21}, hand.Scores())

	// Don't show bust scores if there are other scores that are ok
	hand = Hand{[]*Card{
		NewCard(Ace, Hearts),
		NewCard(Ace, Spades),
	}}
	assert.Equal(t, []int{2, 12}, hand.Scores())

	// Don't show bust scores if there are other scores that are ok
	hand = Hand{[]*Card{
		NewCard(Ace, Hearts),
		NewCard(Ace, Spades),
		NewCard(Ace, Diamonds),
	}}
	assert.Equal(t, []int{3, 13}, hand.Scores())

	// Don't show bust scores if there are other scores that are ok
	hand = Hand{[]*Card{
		NewCard(Ace, Clubs),
		NewCard(Ace, Diamonds),
		NewCard(Ace, Hearts),
		NewCard(Ace, Spades),
	}}
	assert.Equal(t, []int{4, 14}, hand.Scores())

	// Show the minimum bust score if there are only bust scores
	hand = Hand{[]*Card{
		NewCard(Two, Hearts),
		NewCard(Ace, Hearts),
		NewCard(Three, Hearts),
		NewCard(Ace, Spades),
		NewCard(King, Diamonds),
		NewCard(Six, Clubs),
	}}
	assert.Equal(t, []int{23}, hand.Scores())
}
