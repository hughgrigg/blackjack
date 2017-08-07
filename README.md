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

## Todo

 - Drawing the board
 - Dealing
 - Win
 - Loss
 - Bust
 - Blackjack
 - Push
 - Sticking
 - Hitting
 - Betting
 - Doubling down
 - Splitting
 - Insurance
 - Hints
 - Other CLI options, e.g. to ask for a hint for a given situation
