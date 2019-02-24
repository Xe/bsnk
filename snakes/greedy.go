package snakes

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/SolarLune/paths"
	"github.com/Xe/bsnk/api"
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

	g := paths.NewGrid(decoded.Board.Width, decoded.Board.Height)

	for _, sk := range decoded.Board.Snakes {
		for _, pt := range sk.Body {
			c := g.Get(pt.X, pt.Y)
			c.Walkable = false
		}
	}

	path := g.GetPath(g.Get(me[0].X, me[0].Y), g.Get(target.X, target.Y), false)
	if len(path.Cells) != 0 {
		t := path.Next()
		immedTarget := api.Coord{
			X: t.X,
			Y: t.Y,
		}
		pickDir = me[0].Dir(immedTarget)
		ln.Log(ctx, ln.Info("making move"), logCoords("immed_target", immedTarget), ln.F{"pick_dir": pickDir})
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
