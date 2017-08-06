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
	deck.Shuffle(23)
	assert.Equal(t, "J♥", deck.Cards[0].Notation())
	assert.Equal(t, "7♦", deck.Cards[1].Notation())
	assert.Equal(t, "J♤", deck.Cards[2].Notation())
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
