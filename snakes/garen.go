package snakes

import (
	"context"

	"github.com/Xe/bsnk/api"
)

type Garen struct{}

func (g Garen) Start(ctx context.Context, sr api.SnakeRequest) (*api.StartResponse, error) {
	return &api.StartResponse{
		Color: "#FFFF00",
	}, nil
}

func (g Garen) Move(ctx context.Context, sr api.SnakeRequest) (*api.MoveResponse, error) {
	directions := []string{"up", "left", "down", "right"}
	pickDir := directions[sr.Turn%len(directions)]
	return &api.MoveResponse{
		Move: pickDir,
	}, nil
}

func (g Garen) End(ctx context.Context, sr api.SnakeRequest) error {
	return nil
}
