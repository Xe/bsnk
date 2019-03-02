package snakes

import (
	"context"
	"fmt"

	"github.com/Xe/bsnk/api"
	"within.website/ln"
)

// Greedy is a greedy snake AI. It will try to get as long as it can as fast as
// it can.
type Greedy struct{}

// Start starts a game.
func (Greedy) Start(ctx context.Context, gs api.SnakeRequest) (*api.StartResponse, error) {
	return &api.StartResponse{
		Color: "#c79dd7",
	}, nil
}

// Move responds with the snake's movements for a given Turn.
func (g Greedy) Move(ctx context.Context, decoded api.SnakeRequest) (*api.MoveResponse, error) {
	me := decoded.You.Body
	var pickDir string

	_, pf := makePathfinder(decoded)
	target := selectGreedy(decoded)

	ln.WithF(ctx, logCoords("target", target))
	ln.Log(ctx, ln.Info("found_target"))

	path, _ := pf.FindPath(me[0].X, me[0].Y, target.X, target.Y)
	if len(path) >= 2 {
		pickDir = me[0].Dir(api.Coord{
			X: path[1].X,
			Y: path[1].Y,
		})
	} else {
		for _, place := range []api.Coord{me[0].Up(), me[0].Down(), me[0].Left(), me[0].Right()} {
			if !decoded.Board.IsDeadly(place) {
				pickDir = me[0].Dir(place)
			}
		}
	}

	return &api.MoveResponse{
		Move: pickDir,
	}, nil
}

// End ends a game.
func (Greedy) End(ctx context.Context, sr api.SnakeRequest) error {
	return nil
}

func selectGreedy(gs api.SnakeRequest) api.Coord {
	me := gs.You.Body
	var target api.Coord
	var foundTarget bool
	var distance float64 = 99999999999

	for _, fd := range gs.Board.Food {
		l := api.Line{A: me[0], B: fd}
		if sc := l.Manhattan(); sc < distance {
			distance = sc
			target = fd
			foundTarget = true
		}
	}

	if !foundTarget {
		tail := me[len(me)-1]
		target = tail
	}

	return target
}

func logCoords(pfx string, coord api.Coord) ln.F {
	return ln.F{
		pfx + "_x,y": fmt.Sprintf("(%d,%d)", coord.X, coord.Y),
	}
}
