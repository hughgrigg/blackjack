package cards

import (
	"fmt"
	"math/rand"
	"time"
	"bytes"
	"github.com/hughgrigg/blackjack/util"
	"sort"
)

// Suits
type Suit rune

const (
	Clubs    Suit = '♧'
	Diamonds Suit = '♦'
	Hearts   Suit = '♥'
	Spades   Suit = '♤'
)

var Suits = [4]Suit{Clubs, Diamonds, Hearts, Spades}

//
// Ranks
//
type Rank rune

const (
	Ace   Rank = 'A'
	Two   Rank = '2'
	Three Rank = '3'
	Four  Rank = '4'
	Five  Rank = '5'
	Six   Rank = '6'
	Seven Rank = '7'
	Eight Rank = '8'
	Nine  Rank = '9'
	Ten   Rank = 'X'
	Jack  Rank = 'J'
	Queen Rank = 'Q'
	King  Rank = 'K'
)

var Ranks = [13]Rank{
	Ace, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King,
}

var RankValues = map[Rank]int{
	Ace:   1,
	Two:   2,
	Three: 3,
	Four:  4,
	Five:  5,
	Six:   6,
	Seven: 7,
	Eight: 8,
	Nine:  9,
	Ten:   10,
	Jack:  10,
	Queen: 10,
	King:  10,
}

//
// Card
//
type Card struct {
	rank Rank
	suit Suit
}

func (c *Card) Values() []int {
	if c.rank == Ace {
		return []int{1, 11}
	}
	return []int{RankValues[c.rank]}
}

func (c *Card) Notation() string {
	return fmt.Sprintf("%s%s", string(c.rank), string(c.suit))
}

func (c *Card) Render() string {
	if c.suit == Hearts || c.suit == Diamonds {
		// Colour syntax provided by gizak/termui
		return fmt.Sprintf("[%s](fg-red)", c.Notation())
	}
	return c.Notation()
}

//
// Deck
//
type Deck struct {
	Cards []Card
}

func (d *Deck) Init() {
	for _, s := range Suits {
		for _, r := range Ranks {
			d.Cards = append(d.Cards, Card{r, s})
		}
	}
}

const UniqueShuffle = iota

func (d *Deck) Shuffle(seed int64) {
	if seed == UniqueShuffle {
		seed = time.Now().UnixNano()
	}
	rand.Seed(seed)
	for i := 0; i < 52; i++ {
		r := i + rand.Intn(52 - i)
		d.Cards[r], d.Cards[i] = d.Cards[i], d.Cards[r]
	}
}

func (d *Deck) Render() string {
	var buffer bytes.Buffer
	for _, c := range d.Cards {
		buffer.WriteString(c.Render() + " ")
	}
	return buffer.String()
}

//
// Hand
//
type Hand struct {
	Cards []Card
}

func (h *Hand) Hit(c Card) {
	h.Cards = append(h.Cards, c)
}

// Due to aces being 1 or 11, a hand can be worth different scores.
func (h *Hand) Scores() []int {
	scores := []int{0}
	for _, card := range h.Cards {
		for i, score := range scores {
			// Add the first value to each score branch
			scores[i] += card.Values()[0]
			for _, cardValue := range card.Values()[1:] {
				// Make more score branches for further card values
				scores = append(scores, score + cardValue)
			}
		}
	}
	scores = util.UniqueInts(scores)
	sort.Ints(scores)
	return scores
}
