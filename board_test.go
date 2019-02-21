package main

import (
	"fmt"
	"testing"

	"github.com/Xe/bsnk/api"
)

func TestBoardIsInBoard(t *testing.T) {
	cases := []struct {
		width, height int
		coord         api.Coord
		result        bool
	}{
		{
			width:  3,
			height: 3,
			coord:  api.Coord{X: 1, Y: 2},
			result: true,
		},
		{
			width:  3,
			height: 3,
			coord:  api.Coord{X: -1, Y: 2},
			result: false,
		},
		{
			width:  3,
			height: 3,
			coord:  api.Coord{X: 1, Y: -2},
			result: false,
		},
		{
			width:  3,
			height: 3,
			coord:  api.Coord{X: 4, Y: 2},
			result: false,
		},
		{
			width:  3,
			height: 3,
			coord:  api.Coord{X: 2, Y: 4},
			result: false,
		},
	}

	for _, cs := range cases {
		t.Run(fmt.Sprintf("%dx%d - %#v", cs.width, cs.height, cs.coord), func(t *testing.T) {
			b := Board{
				Board: api.Board{
					Height: cs.height,
					Width:  cs.width,
				},
			}

			got := b.isInBoard(cs.coord)
			if cs.result != got {
				t.Errorf("wanted (%d,%d) to be inside a %dx%x grid: %v, got: %v", cs.coord.X, cs.coord.Y, cs.height, cs.width, cs.result, got)
			}
		})
	}
}
