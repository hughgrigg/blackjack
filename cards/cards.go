package cards

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/hughgrigg/blackjack/util"
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
	show bool
}

func NewCard(rank Rank, suit Suit) *Card {
	return &Card{rank, suit, true}
}

func (c *Card) Values() []int {
	if c.rank == Ace {
		return []int{1, 11}
	}
	return []int{RankValues[c.rank]}
}

func (c *Card) Notation() string {
	if c.show {
		return fmt.Sprintf("%s%s", string(c.rank), string(c.suit))
	}
	return "🂠 ?"
}

func (c *Card) FaceDown() *Card {
	c.show = false
	return c
}

func (c *Card) FaceUp() *Card {
	c.show = true
	return c
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
	Cards []*Card
}

func (d *Deck) Init() {
	for _, s := range Suits {
		for _, r := range Ranks {
			d.Cards = append(d.Cards, &Card{r, s, true})
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
		r := i + rand.Intn(52-i)
		d.Cards[r], d.Cards[i] = d.Cards[i], d.Cards[r]
	}
}

func (d *Deck) Pop() *Card {
	card := *d.Cards[len(d.Cards)-1]
	d.Cards = d.Cards[:len(d.Cards)-1]
	return &card
}

func (d Deck) Render() string {
	return fmt.Sprintf("🂠  ×%d", len(d.Cards))
}

//
// Hands
//
type Hand struct {
	Cards []*Card
}

func (h *Hand) Hit(c *Card) {
	if h.Cards == nil {
		h.Cards = []*Card{}
	}
	h.Cards = append(h.Cards, c)
}

func (h *Hand) IsBust() bool {
	return util.MinInt(h.Scores()) > 21
}

func (h *Hand) HasBlackJack() bool {
	for _, score := range h.Scores() {
		if score == 21 {
			return true
		}
	}
	return false
}

func (h Hand) Render() string {
	buffer := bytes.Buffer{}
	last := len(h.Cards) - 1
	for i, card := range h.Cards {
		buffer.WriteString(card.Render())
		if i != last {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(fmt.Sprintf(
		"  (%s)", scoresRenderer{h.Scores()}.Render(),
	))
	return buffer.String()
}

// Player can have many hands
type HandSet struct {
	Hands []*Hand
}

func (hs HandSet) Render() string {
	buffer := bytes.Buffer{}
	last := len(hs.Hands) - 1
	for i, hand := range hs.Hands {
		buffer.WriteString(hand.Render())
		if i != last {
			buffer.WriteString(" | ")
		}
	}
	return buffer.String()
}

// Due to aces being 1 or 11, a hand can be worth different scores.
func (h *Hand) Scores() []int {
	scores := []int{0}
	for _, card := range h.Cards {
		if !card.show {
			continue
		}
		for i, score := range scores {
			// Add the first value to each score branch
			scores[i] += card.Values()[0]
			for _, cardValue := range card.Values()[1:] {
				// Make more score branches for further card values
				scores = append(scores, score+cardValue)
			}
		}
	}
	// If we have blackjack then return that alone
	for _, score := range scores {
		if score == 21 {
			return []int{21}
		}
	}
	scores = util.UniqueInts(scores)
	scores = sanitiseScores(scores)
	sort.Ints(scores)
	return scores
}

func sanitiseScores(scores []int) []int {
	// Don't give bust scores if there are other scores that are ok
	minScore := util.MinInt(scores)
	if minScore <= 21 {
		okScores := []int{}
		for _, score := range scores {
			if score <= 21 {
				okScores = append(okScores, score)
			}
		}
		return okScores
	}
	// Give the minimum bust score if there are only bust scores
	return []int{minScore}
}

type scoresRenderer struct {
	scores []int
}

func (sr scoresRenderer) Render() string {
	buffer := bytes.Buffer{}
	last := len(sr.scores) - 1
	for i, score := range sr.scores {
		buffer.WriteString(fmt.Sprintf("%d", score))
		if i != last {
			buffer.WriteString(" / ")
		}
	}
	return buffer.String()
}
