package main

import (
	"testing"

	"github.com/Xe/bsnk/api"
)

func TestSelectTarget(t *testing.T) {
	cases := []struct {
		name          string
		data          api.SnakeRequest
		target, immed api.Coord
	}{
		{
			name: "right",
			data: api.SnakeRequest{
				Game: api.Game{
					ID: "right",
				},
				Board: api.Board{
					Width:  11,
					Height: 11,
					Food: []api.Coord{
						{
							X: 4,
							Y: 7,
						},
					},
				},
				You: api.Snake{
					Body: []api.Coord{
						{X: 1, Y: 1},
						{X: 1, Y: 1},
						{X: 1, Y: 1},
					},
				},
			},
			target: api.Coord{
				X: 4,
				Y: 7,
			},
			immed: api.Coord{
				X: 2,
				Y: 1,
			},
		},
		{
			name: "down",
			data: api.SnakeRequest{
				Game: api.Game{
					ID: "down",
				},
				Board: api.Board{
					Width:  11,
					Height: 11,
					Food: []api.Coord{
						{
							X: 1,
							Y: 7,
						},
					},
				},
				You: api.Snake{
					Body: []api.Coord{
						{X: 1, Y: 1},
						{X: 1, Y: 1},
						{X: 1, Y: 1},
					},
				},
			},
			target: api.Coord{
				X: 1,
				Y: 7,
			},
			immed: api.Coord{
				X: 1,
				Y: 2,
			},
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			target, immed := selectTarget(cs.data)

			if !target.Eq(cs.target) {
				t.Errorf("wanted target: %s, got: %s", cs.target, target)
			}

			if !immed.Eq(cs.immed) {
				t.Errorf("wanted immed: %s, got: %s", cs.immed, immed)
			}
		})
	}
}
