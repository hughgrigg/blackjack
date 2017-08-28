package cards

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"math/big"

	"strings"

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

// These are initial values of cards. Card.Values() should be used to get its
// possible values.
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
	rank   Rank
	suit   Suit
	faceUp bool
}

// Construct a card with a rank and suit.
func NewCard(rank Rank, suit Suit) *Card {
	return &Card{rank, suit, true}
}

// Get the possible values of a card. Ace being either 1 or 11 is generalised
// to all cards' values being a slice of ints.
func (c *Card) Values() []int {
	if c.rank == Ace {
		return []int{1, 11}
	}
	return []int{RankValues[c.rank]}
}

// Get a plain string notation for the card, e.g. A♤ for the Ace of Spades.
func (c *Card) Notation() string {
	if c.faceUp {
		return fmt.Sprintf("%s%s", string(c.rank), string(c.suit))
	}
	return "🂠 ?"
}

// IsFaceUp sees if a card is facing up.
func (c *Card) IsFaceUp() bool {
	return c.faceUp
}

// FaceDown turns a card face down so its true notation is not displayed.
func (c *Card) FaceDown() *Card {
	c.faceUp = false
	return c
}

// FaceUp turns a card face up so its notation is visible.
func (c *Card) FaceUp() *Card {
	c.faceUp = true
	return c
}

// Get a colour-coded rendering of the card as a string.
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

// Initialise the deck with all 52 cards in order.
func (d *Deck) Init() {
	d.Cards = []*Card{}
	for _, s := range Suits {
		for _, r := range Ranks {
			d.Cards = append(d.Cards, &Card{r, s, true})
		}
	}
}

const UniqueShuffle = iota

// Shuffle the deck to an order based on a seed value. UniqueShuffle can be
// passed to get random shuffling.
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

// Pop the top card off the deck.
func (d *Deck) Pop() *Card {
	card := *d.Cards[len(d.Cards)-1]
	d.Cards = d.Cards[:len(d.Cards)-1]
	return &card
}

// Ensure the next card to be popped is a specific card.
func (d *Deck) ForceNext(c *Card) {
	existing := -1
	for i, card := range d.Cards {
		if card == c {
			existing = i
		}
	}
	last := len(d.Cards) - 1
	if existing > -1 {
		d.Cards[last], d.Cards[existing] = d.Cards[existing], d.Cards[last]
	} else {
		d.Cards = append(d.Cards, c)
	}
}

// Get a rendering of the deck as a string.
func (d Deck) Render() string {
	return fmt.Sprintf("🂠  ×%d", len(d.Cards))
}

//
// Hands
//
type Hand struct {
	Cards []*Card
}

// Add a card to the hand.
func (h *Hand) Hit(c *Card) {
	if h.Cards == nil {
		h.Cards = []*Card{}
	}
	h.Cards = append(h.Cards, c)
}

// IsBust sees if the hand is bust, i.e. it has no possible scores less than 22.
func (h *Hand) IsBust() bool {
	return util.MinInt(h.Scores()) > 21
}

// HasBlackJack sees if the hand has blackjack, i.e. 21 is one of its possible
// scores.
func (h *Hand) HasBlackJack() bool {
	// Blackjack is achieved with 2 cards only, otherwise it's just 21.
	if len(h.Cards) != 2 {
		return false
	}
	for _, score := range h.Scores() {
		if score == 21 {
			return true
		}
	}
	return false
}

// IsSoft sees if a hand has soft scores, i.e. based on an ace being 1 or 11.
func (h *Hand) IsSoft() bool {
	// A bust hand can't be soft.
	if h.IsBust() {
		return false
	}
	for _, card := range h.Cards {
		if card.rank == Ace {
			return true
		}
	}
	return false
}

// HasHard17 sees if a hand has a hard 17 or greater. Dealers hit until hard 17
// or higher during their turn.
func (h *Hand) HasHard17() bool {
	for _, score := range h.Scores() {
		// Any score over 17 is ok.
		if score > 17 {
			return true
		}
		// A score of exactly 17 must be a hard 17.
		if score == 17 && !h.IsSoft() {
			return true
		}
	}
	return false
}

// WinFactor assesses whether one hand beats another, giving the multiplier for
// calculating the winnings. E.g. 2.5 for blackjack, 2 for winning, 1 for push
// and 0 for losing.
func (h *Hand) WinFactor(other *Hand) *big.Float {
	ourScore := util.MaxInt(h.Scores())
	theirScore := util.MaxInt(other.Scores())

	// Push if we got the same score.
	if ourScore == theirScore {
		return big.NewFloat(1)
	}

	// Lose if we bust.
	if h.IsBust() {
		return big.NewFloat(0)
	}

	// Extra points if we have blackjack.
	if h.HasBlackJack() {
		return big.NewFloat(2.5)
	}

	// Win if they bust.
	if other.IsBust() {
		return big.NewFloat(2)
	}

	// Win if we beat their score.
	if ourScore > theirScore {
		return big.NewFloat(2)
	}

	// Otherwise we've lost.
	return big.NewFloat(0)
}

// CanSplit is true if the hand can legally be split into two hands. This is
// allowed when the hand contains two cards of the same rank.
func (h *Hand) CanSplit() bool {
	if len(h.Cards) != 2 {
		return false
	}
	for _, first := range h.Cards[0].Values() {
		if !util.IntsContain(first, h.Cards[1].Values()) {
			return false
		}
	}
	return true
}

// Render the hand as a string.
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
	return strings.Trim(buffer.String(), " ")
}

// Get the possible scores for the hand. Due to aces being 1 or 11, a hand can
// be worth different scores.
func (h *Hand) Scores() []int {
	scores := []int{0}
	for _, card := range h.Cards {
		if !card.faceUp {
			continue
		}
		for i, score := range scores {
			// Add the first value to each score branch.
			scores[i] += card.Values()[0]
			for _, cardValue := range card.Values()[1:] {
				// Make more score branches for further card values.
				// In other words, branch on aces.
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

// Make a set of hand scores more human readable.
func sanitiseScores(scores []int) []int {
	// Don't give bust scores if there are other scores that are ok.
	// E.g. we don't care that 27 is possible if 17 is also possible.
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
	// Give the minimum bust score if there are only bust scores.
	// E.g. knowing we bust on 25 is enough, we don't need to see a 35 as well.
	return []int{minScore}
}

type scoresRenderer struct {
	scores []int
}

// Get a rendering of the scores as a string.
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
