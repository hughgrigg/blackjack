package main

import (
	"github.com/gizak/termui"
	"time"
	"github.com/hughgrigg/blackjack/ui"
)

func main() {
	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	display := ui.Display{}
	display.Init()

	display.Render()
	go func() {
		for range time.Tick(time.Millisecond * 100) {
			display.Render()
		}
	}()

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})
	termui.Loop()
}
