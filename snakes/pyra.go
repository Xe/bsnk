package snakes

import (
	"context"

	"github.com/Xe/bsnk/api"
	"github.com/prettymuchbryce/goeasystar"
	"within.website/ln"
	"within.website/ln/opname"
)

// Pyra is a slightly smarter version of Greedy.
//
// Struct memebers are configuration flags for the snake behavior.
type Pyra struct {
	MinLength int

	targets map[string]pyraState
}

type pyraTarget struct {
	api.Line

	Score       int
	AstarLength int
}

type pyraState struct {
	path []*goeasystar.Point
	trg  *pyraTarget
}

func (pt pyraTarget) F() ln.F {
	f := ln.F{
		"target_score":        pt.Score,
		"target_astar_length": pt.AstarLength,
	}

	f.Extend(logCoords("my_head", pt.Line.A))
	f.Extend(logCoords("target_coords", pt.Line.B))

	return f
}

// Start starts a game.
func (p *Pyra) Start(ctx context.Context, gs api.SnakeRequest) (*api.StartResponse, error) {
	if p.targets == nil {
		p.targets = map[string]pyraState{}
	}

	p.targets[gs.Game.ID] = p.getState(ctx, gs)

	return &api.StartResponse{
		Color:    "#5ce8c3",
		HeadType: "beluga",
		TailType: "skinny",
	}, nil
}

func (p *Pyra) getState(ctx context.Context, sr api.SnakeRequest) pyraState {
	me := sr.You.Body

	_, pf := makePathfinder(sr)
	target := p.selectTarget(ctx, sr, pf)

	path, _ := pf.FindPath(me[0].X, me[0].Y, target.Line.B.X, target.Line.B.Y)

	return pyraState{
		path: path,
		trg:  &target,
	}
}

// Move responds with the snake's movements for a given Turn.
func (p *Pyra) Move(ctx context.Context, decoded api.SnakeRequest) (*api.MoveResponse, error) {
	me := decoded.You.Body
	var pickDir string

	st := p.targets[decoded.Game.ID]

	if len(st.path) > 2 {
		st = p.getState(ctx, decoded)
	}

	if len(st.path) > 2 {
		for _, coord := range []api.Coord{me[0].Up(), me[0].Left(), me[0].Right(), me[0].Down()} {
			if decoded.Board.Inside(coord) && !decoded.Board.IsDeadly(coord) {
				pickDir = me[0].Dir(coord)
				break
			}
		}
	} else {
		pickDir = me[0].Dir(api.Coord{
			X: st.path[1].X,
			Y: st.path[1].Y,
		})
		st.path = st.path[1:]
	}

	p.targets[decoded.Game.ID] = st

	return &api.MoveResponse{
		Move: pickDir,
	}, nil
}

// End ends a game.
func (p *Pyra) End(ctx context.Context, sr api.SnakeRequest) error {
	delete(p.targets, sr.Game.ID)

	return nil
}

func (p Pyra) selectTarget(ctx context.Context, gs api.SnakeRequest, pf *goeasystar.Pathfinder) pyraTarget {
	ctx = opname.With(ctx, "select-target")
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
		for _, place := range []api.Coord{tail.Up(), tail.Down(), tail.Left(), tail.Right()} {
			path, err := pf.FindPath(me[0].X, me[0].Y, place.X, place.Y)
			if err != nil {
				continue
			}

			targets = append(targets, pyraTarget{
				Line: api.Line{
					A: me[0],
					B: tail,
				},
				Score:       50,
				AstarLength: len(path),
			})
			break
		}
	}

	for _, sn := range gs.Board.Snakes {
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
		ln.Log(ctx, ln.Info("no targets found"))
		for _, place := range []api.Coord{me[0].Up(), me[0].Down(), me[0].Left(), me[0].Right()} {
			if !gs.Board.IsDeadly(place) {
				return pyraTarget{
					Line: api.Line{
						A: me[0],
						B: place,
					},
				}
			}
		}
	}

	var t pyraTarget
	for _, pt := range targets {
		pt.Score = pt.Score - int(pt.Line.Manhattan())
		for _, place := range []api.Coord{pt.Line.B.Up(), pt.Line.B.Down(), pt.Line.B.Left(), pt.Line.B.Right()} {
			if gs.Board.IsDeadly(place) {
				goto next
			}
		}

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
	next:
	}

	ln.Log(ctx, ln.Info("found target"), t)

	return t
}
