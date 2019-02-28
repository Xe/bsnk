package snakes

import (
	"context"

	"github.com/Xe/bsnk/api"
)

// Erratic is a particularly terrible AI.
type Erratic struct{}

// Start starts a game.
func (Erratic) Start(ctx context.Context, gs api.SnakeRequest) (*api.StartResponse, error) {
	return &api.StartResponse{
		Color: "#7FF3CF",
	}, nil
}

// Move twitches around.
func (Erratic) Move(ctx context.Context, gs api.SnakeRequest) (*api.MoveResponse, error) {
	me := gs.You.Body
	var pickDir string

	for place := range map[api.Coord]struct{}{
		me[0].Up():    struct{}{},
		me[0].Down():  struct{}{},
		me[0].Left():  struct{}{},
		me[0].Right(): struct{}{},
	} {
		if gs.Board.Inside(place) && !gs.Board.IsDeadly(place) {
			pickDir = me[0].Dir(place)
		}
	}

	return &api.MoveResponse{
		Move: pickDir,
	}, nil
}

// End ends a game.
func (Erratic) End(ctx context.Context, sr api.SnakeRequest) error {
	return nil
}
