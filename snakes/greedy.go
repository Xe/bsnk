package snakes

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/Xe/bsnk/api"
	"github.com/prettymuchbryce/goeasystar"
	"within.website/ln"
)

type Greedy struct{}

func (Greedy) Start(ctx context.Context, gs api.SnakeRequest) (*api.StartResponse, error) {
	return &api.StartResponse{
		Color: "#c79dd7",
	}, nil
}

func (Greedy) Move(ctx context.Context, decoded api.SnakeRequest) (*api.MoveResponse, error) {
	me := decoded.You.Body
	var pickDir string

	target := selectFood(decoded)
	if len(decoded.Board.Food) == 0 {
		target = me[len(me)-1]
	}

	ln.WithF(ctx, logCoords("target", target))
	ln.Log(ctx, ln.Info("found_target"))

	pf := goeasystar.NewPathfinder()
	pf.DisableCornerCutting()
	pf.DisableDiagonals()
	pf.SetAcceptableTiles([]int{1})

	var grid [][]int
	grid = make([][]int, decoded.Board.Height)
	for i := range grid {
		grid[i] = make([]int, decoded.Board.Width)
		for j := range grid[i] {
			grid[i][j] = 1
		}
	}

	pf.SetGrid(grid)

	for _, sk := range decoded.Board.Snakes {
		for _, pt := range sk.Body {
			grid[pt.X][pt.Y] = 2
			pf.AvoidAdditionalPoint(pt.X, pt.Y)
		}
	}

	path, err := pf.FindPath(me[0].X, me[0].Y, target.X, target.Y)
	if err != nil {
		for _, place := range []api.Coord{me[0].Up(), me[0].Left(), me[0].Down(), me[0].Right()} {
			if !decoded.Board.IsDeadly(place) {
				pickDir = me[0].Dir(place)
			}
		}
	} else {
		pickDir = me[0].Dir(api.Coord{
			X: path[1].X,
			Y: path[1].Y,
		})
	}

	return &api.MoveResponse{
		Move: pickDir,
	}, nil
}

func (Greedy) End(ctx context.Context, sr api.SnakeRequest) error {
	return nil
}

func manhattan(l, r api.Coord) float64 {
	absX := r.X - l.X
	if absX < 0 {
		absX = -absX
	}

	absY := r.Y - l.Y
	if absY < 0 {
		absY = -absY
	}

	return float64(absX + absY)
}

func selectFood(gs api.SnakeRequest) api.Coord {
	me := gs.You.Body
	var target api.Coord
	var foundTarget bool
	var distance float64 = 99999999999

	for _, fd := range gs.Board.Food {
		if sc := manhattan(me[0], fd); sc < distance {
			distance = sc
			target = fd
			foundTarget = true
		}
	}

	if !foundTarget {
		target = api.Coord{
			X: rand.Intn(gs.Board.Width),
			Y: rand.Intn(gs.Board.Height),
		}
	}

	return target
}

func logCoords(pfx string, coord api.Coord) ln.F {
	return ln.F{
		pfx + "_x,y": fmt.Sprintf("(%d,%d)", coord.X, coord.Y),
	}
}
