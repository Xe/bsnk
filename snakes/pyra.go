package snakes

import (
	"context"

	"github.com/Xe/bsnk/api"
	"github.com/prettymuchbryce/goeasystar"
	"within.website/ln"
)

// Pyra is a slightly smarter version of Greedy.
//
// Struct memebers are configuration flags for the snake behavior.
type Pyra struct {
	MinLength int
}

type pyraTarget struct {
	api.Line

	Score       int
	AstarLength int
}

// Start starts a game.
func (Pyra) Start(ctx context.Context, gs api.SnakeRequest) (*api.StartResponse, error) {
	return &api.StartResponse{
		Color: "#c79dd7",
	}, nil
}

// Move responds with the snake's movements for a given Turn.
func (p Pyra) Move(ctx context.Context, decoded api.SnakeRequest) (*api.MoveResponse, error) {
	me := decoded.You.Body
	var pickDir string

	pf := goeasystar.NewPathfinder()
	pf.DisableCornerCutting()
	pf.DisableDiagonals()
	pf.SetAcceptableTiles([]int{1, 2, 5, 8})

	var grid [][]int
	grid = make([][]int, decoded.Board.Height)
	for i := range grid {
		grid[i] = make([]int, decoded.Board.Width)

		for j := range grid[i] {
			if j == 0 || j == len(grid[i])-1 {
				grid[i][j] = 8
			}

			if i == 0 || i == len(grid)-1 {
				grid[i][j] = 8
			} else {
				grid[i][j] = 1
			}
		}
	}

	pf.SetGrid(grid)

	for _, sk := range decoded.Board.Snakes {
		for _, pt := range sk.Body {
			pf.AvoidAdditionalPoint(pt.X, pt.Y)

			if sk.ID != decoded.You.ID {
				for _, st := range []api.Coord{
					pt.Up(),
					pt.Left(),
					pt.Right(),
					pt.Down(),
				} {
					pf.SetAdditionalPointCost(st.X, st.Y, 5)
				}
			}
		}
	}

	target := p.selectTarget(decoded, pf)
	ln.WithF(ctx, logCoords("target", target))
	ln.Log(ctx, ln.Info("found_target"))

	path, _ := pf.FindPath(me[0].X, me[0].Y, target.X, target.Y)
	pickDir = me[0].Dir(api.Coord{
		X: path[1].X,
		Y: path[1].Y,
	})

	return &api.MoveResponse{
		Move: pickDir,
	}, nil
}

// End ends a game.
func (Pyra) End(ctx context.Context, sr api.SnakeRequest) error {
	return nil
}

func (p Pyra) selectTarget(gs api.SnakeRequest, pf *goeasystar.Pathfinder) api.Coord {
	me := gs.You.Body
	var targets []pyraTarget
	for _, fd := range gs.Board.Food {
		t := pyraTarget{
			Line: api.Line{
				A: me[0],
				B: fd,
			},
			Score: 20,
		}

		if len(me) < p.MinLength {
			t.Score = 50
		}

		if gs.You.Health <= 30 {
			t.Score = 9000
		}

		path, err := pf.FindPath(me[0].X, me[0].Y, fd.X, fd.Y)
		if err != nil {
			continue
		}
		t.AstarLength = len(path)

		targets = append(targets, t)
	}

	{
		tail := me[len(me)-1]
		path, err := pf.FindPath(me[0].X, me[0].Y, tail.X, tail.Y)
		if err != nil {
			goto skip
		}

		targets = append(targets, pyraTarget{
			Line: api.Line{
				A: me[0],
				B: tail,
			},
			Score:       30,
			AstarLength: len(path),
		})
	}
skip:

	for _, sn := range gs.Board.Snakes {
		if sn.ID == gs.You.ID {
			continue
		}

		if len(gs.You.Body) < len(sn.Body) {
			continue
		}

		head := sn.Body[0]
		path, err := pf.FindPath(me[0].X, me[0].Y, head.X, head.Y)
		if err != nil {
			continue
		}

		targets = append(targets, pyraTarget{
			Line: api.Line{
				A: me[0],
				B: head,
			},
			Score:       400,
			AstarLength: len(path),
		})
	}

	if len(targets) == 0 {
		for _, place := range []api.Coord{me[0].Up(), me[0].Down(), me[0].Left(), me[0].Right()} {
			if !gs.Board.IsDeadly(place) {
				return place
			}
		}
	}

	var t pyraTarget
	for _, pt := range targets {
		if pt.Score > t.Score {
			// not possible unless t is uninitialized
			if t.AstarLength == 0 {
				t = pt
				continue
			}

			if pt.AstarLength < t.AstarLength {
				t = pt
			}
		}
	}

	return t.Line.B
}
