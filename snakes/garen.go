package snakes

import (
	"context"

	"github.com/Xe/bsnk/api"
)

// Garen spins to win.
type Garen struct{}

func (Garen) Ping() (*api.PingResponse, error) {
	return &api.PingResponse{
		APIVersion: "1",
		Color:      "#FFFF00",
	}, nil
}

// Start kicks off a game.
func (Garen) Start(ctx context.Context, sr api.SnakeRequest) error {
	return nil
}

// Move spins to win.
func (Garen) Move(ctx context.Context, sr api.SnakeRequest) (*api.MoveResponse, error) {
	directions := []string{"up", "left", "down", "right"}
	pickDir := directions[sr.Turn%len(directions)]
	return &api.MoveResponse{
		Move: pickDir,
	}, nil
}

// End ends a game.
func (Garen) End(ctx context.Context, sr api.SnakeRequest) error {
	return nil
}
