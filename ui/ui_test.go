package ui

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDisplay_NewView(t *testing.T) {
	display := Display{}
	firstCount := len(display.views)
	display.NewView("foobar", 5)
	assert.Equal(t, firstCount+1, len(display.views))
}
