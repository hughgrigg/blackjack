Blackjack
=========

[![Build Status](https://travis-ci.org/hughgrigg/blackjack.svg?branch=master)](https://travis-ci.org/hughgrigg/blackjack)
[![Go Report Card](https://goreportcard.com/badge/github.com/hughgrigg/blackjack)](https://goreportcard.com/report/github.com/hughgrigg/blackjack)
[![Coverage Status](https://coveralls.io/repos/github/hughgrigg/blackjack/badge.svg?branch=master)](https://coveralls.io/github/hughgrigg/blackjack?branch=master)

This is an implementation of the
[blackjack](https://en.wikipedia.org/wiki/Blackjack) casino game in Go. This
game is also known as twenty-one.

The game uses a single
[52-card deck](https://en.wikipedia.org/wiki/Standard_52-card_deck).

## Tests

You can run all the tests with:

```bash
go test -v ./...
```

Or for nicer output:

```bash
go get -u github.com/kyoh86/richgo

```

## Todo

 - Dealer stage behaviour (i.e. hit to 17)
 - Refactor "bets and balance" into "bank"
 - Refactor the player's indexed hands into bets being linked to hands
 - Win
 - Loss
 - Push
 - Doubling down
 - Splitting
 - Insurance
 - Hints
 - Other CLI options, e.g. to ask for a hint for a given situation
