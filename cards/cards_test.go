package cards

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

//
// Card
//

func TestNewCard(t *testing.T) {

}

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

func TestCard_FaceUp(t *testing.T) {
	card := NewCard(Ace, Spades)
	card.FaceUp()
	assert.Equal(t, "A♤", card.Notation())
}

func TestCard_FaceDown(t *testing.T) {
	card := NewCard(Ace, Spades)
	card.FaceDown()
	assert.Equal(t, "🂠 ?", card.Notation())
}

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
	assert.Equal(t, "🂠  ×52", deck.Render())
}

func TestDeck_ShuffleFixed(t *testing.T) {
	deck := Deck{}
	deck.Init()
	deck.Shuffle(42)
	assert.Equal(t, "8♥", deck.Cards[0].Notation())
	assert.Equal(t, "9♦", deck.Cards[1].Notation())
	assert.Equal(t, "8♦", deck.Cards[2].Notation())
}

func TestDeck_Pop(t *testing.T) {
	deck := Deck{}
	deck.Init()
	deck.Shuffle(42)
	assert.Equal(t, "5♤", deck.Pop().Notation())
	assert.Equal(t, "Q♤", deck.Pop().Notation())
}

//
// Hand
//

func TestHand_Hit(t *testing.T) {
	hand := Hand{}
	assert.Empty(t, hand.Cards)

	hand.Hit(NewCard(Ace, Spades))
	hand.Hit(NewCard(Jack, Diamonds))
	assert.Len(t, hand.Cards, 2)
	assert.Equal(t, NewCard(Ace, Spades), hand.Cards[0])
	assert.Equal(t, NewCard(Jack, Diamonds), hand.Cards[1])
}

func TestHand_Render(t *testing.T) {
	hand := Hand{}
	hand.Hit(NewCard(Ace, Spades))
	hand.Hit(NewCard(Jack, Clubs))
	assert.Equal(t, "A♤, J♧", hand.Render())
}

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

	hand = Hand{[]*Card{
		NewCard(Ace, Hearts),
		NewCard(Ace, Spades),
	}}
	assert.Equal(t, []int{2, 12, 22}, hand.Scores())

	hand = Hand{[]*Card{
		NewCard(Ace, Hearts),
		NewCard(Ace, Spades),
		NewCard(Ace, Diamonds),
	}}
	assert.Equal(t, []int{3, 13, 23, 33}, hand.Scores())

	hand = Hand{[]*Card{
		NewCard(Ace, Clubs),
		NewCard(Ace, Diamonds),
		NewCard(Ace, Hearts),
		NewCard(Ace, Spades),
	}}
	assert.Equal(t, []int{4, 14, 24, 34, 44}, hand.Scores())

	hand = Hand{[]*Card{
		NewCard(Two, Hearts),
		NewCard(Ace, Hearts),
		NewCard(Three, Hearts),
		NewCard(Ace, Spades),
		NewCard(King, Diamonds),
		NewCard(Six, Clubs),
	}}
	assert.Equal(t, []int{23, 33, 43}, hand.Scores())
}
