package game

import (
	"testing"
	"time"

	"github.com/hughgrigg/blackjack/cards"
	"github.com/stretchr/testify/assert"
)

func TestBoard_Begin(t *testing.T) {
	board := Board{}
	board.Begin(0)
	time.Sleep(10 * time.Millisecond) // let actions clear
	assert.NotEmpty(t, board.Deck.Cards)
	assert.NotEmpty(t, board.Dealer.Render())
}

func TestDealer_Render(t *testing.T) {
	dealer := Dealer{&cards.Hand{}}
	dealer.hand.Hit(cards.NewCard(cards.Ace, cards.Spades))
	assert.Equal(t, "A♤", dealer.Render())
}
