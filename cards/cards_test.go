package cards

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

//
// Card
//

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
		card := Card{rank, Spades}
		assert.Equal(t, values, card.Values())
	}
}

func TestCard_Notation(t *testing.T) {
	expected := map[Card]string{
		{Ace, Spades}:     "A♤",
		{Queen, Hearts}:   "Q♥",
		{Two, Clubs}:      "2♧",
		{Eight, Diamonds}: "8♦",
	}
	for card, drawn := range expected {
		assert.Equal(t, drawn, card.Notation())
	}
}

func TestCard_Render(t *testing.T) {
	expected := map[Card]string{
		{Ten, Clubs}:     "X♧",
		{King, Spades}:   "K♤",
		{Five, Hearts}:   "[5♥](fg-red)",
		{Jack, Diamonds}: "[J♦](fg-red)",
	}
	for card, drawn := range expected {
		assert.Equal(t, drawn, card.Render())
	}
}

//
// Deck
//

func TestDeck_Init(t *testing.T) {
	deck := Deck{}
	deck.Init()
	assert.Equal(t, "A♧", deck.Cards[00].Notation())
	assert.Equal(t, "A♦", deck.Cards[13].Notation())
	assert.Equal(t, "A♥", deck.Cards[26].Notation())
	assert.Equal(t, "A♤", deck.Cards[39].Notation())
}

func TestDeck_Render(t *testing.T) {
	deck := Deck{}
	deck.Init()
	assert.Equal(t, "A♧ 2♧ 3♧ ", deck.Render()[0:15])
}

func TestDeck_ShuffleFixed(t *testing.T) {
	deck := Deck{}
	deck.Init()
	deck.Shuffle(42)
	assert.Equal(t, "8♥", deck.Cards[0].Notation())
	assert.Equal(t, "9♦", deck.Cards[1].Notation())
	assert.Equal(t, "8♦", deck.Cards[2].Notation())
}

func TestDeck_ShuffleReal(t *testing.T) {
	deck := Deck{}
	deck.Init()

	deck.Shuffle(UniqueShuffle)
	orderA := deck.Render()
	deck.Shuffle(UniqueShuffle)
	orderB := deck.Render()

	assert.NotEqual(t, orderA, orderB)
}

func TestHand_Hit(t *testing.T) {
	hand := Hand{}
	assert.Empty(t, hand.Cards)

	hand.Hit(Card{Ace, Spades})
	hand.Hit(Card{Jack, Diamonds})
	assert.Len(t, hand.Cards, 2)
	assert.Equal(t, Card{Ace, Spades}, hand.Cards[0])
	assert.Equal(t, Card{Jack, Diamonds}, hand.Cards[1])
}

func TestHand_Scores(t *testing.T) {
	var hand Hand

	hand = Hand{}
	assert.Equal(t, []int{0}, hand.Scores())

	hand = Hand{[]Card{
		{Five, Hearts},
	}}
	assert.Equal(t, []int{5}, hand.Scores())

	hand = Hand{[]Card{
		{Five, Hearts},
		{Three, Hearts},
	}}
	assert.Equal(t, []int{8}, hand.Scores())

	hand = Hand{[]Card{
		{Jack, Hearts},
		{Queen, Hearts},
		{King, Hearts},
	}}
	assert.Equal(t, []int{30}, hand.Scores())

	hand = Hand{[]Card{
		{Five, Hearts},
		{Ace, Hearts},
	}}
	assert.Equal(t, []int{6, 16}, hand.Scores())

	hand = Hand{[]Card{
		{Five, Hearts},
		{Ace, Hearts},
		{Three, Hearts},
	}}
	assert.Equal(t, []int{9, 19}, hand.Scores())

	hand = Hand{[]Card{
		{Ace, Hearts},
	}}
	assert.Equal(t, []int{1, 11}, hand.Scores())

	hand = Hand{[]Card{
		{Ace, Hearts},
		{Ace, Spades},
	}}
	assert.Equal(t, []int{2, 12, 22}, hand.Scores())

	hand = Hand{[]Card{
		{Ace, Hearts},
		{Ace, Spades},
		{Ace, Diamonds},
	}}
	assert.Equal(t, []int{3, 13, 23, 33}, hand.Scores())

	hand = Hand{[]Card{
		{Ace, Clubs},
		{Ace, Diamonds},
		{Ace, Hearts},
		{Ace, Spades},
	}}
	assert.Equal(t, []int{4, 14, 24, 34, 44}, hand.Scores())

	hand = Hand{[]Card{
		{Two, Hearts},
		{Ace, Hearts},
		{Three, Hearts},
		{Ace, Spades},
		{King, Diamonds},
		{Six, Clubs},
	}}
	assert.Equal(t, []int{23, 33, 43}, hand.Scores())
}
