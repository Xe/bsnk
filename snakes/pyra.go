package snakes

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/Xe/bsnk/api"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prettymuchbryce/goeasystar"
	"within.website/ln"
	"within.website/ln/opname"
)

// Pyra is a slightly smarter version of Greedy.
//
// Struct memebers are configuration flags for the snake behavior.
type Pyra struct {
	Redis     *redis.Client
	MinLength int
}

type pyraTarget struct {
	api.Line

	Score       int
	AstarLength int
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

var (
        pyraGamesStarted = promauto.NewCounter(prometheus.CounterOpts{
                Name: "pyra_games_started",
                Help: "The number of games started",
        })

	pyraMovesMade = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pyra_moves_made",
		Help: "The number of moves made",
	})

	pyraGamesEnded = promauto.NewCounter(prometheus.CounterOpts{
                Name: "pyra_games_ended",
                Help: "The number of games ended",
        })
)


// Start starts a game.
func (Pyra) Start(ctx context.Context, gs api.SnakeRequest) (*api.StartResponse, error) {
	pyraGamesStarted.Inc()
	return &api.StartResponse{
		Color:    "#FFD600",
		HeadType: "beluga",
		TailType: "skinny",
	}, nil
}

// Move responds with the snake's movements for a given Turn.
func (p Pyra) Move(ctx context.Context, decoded api.SnakeRequest) (*api.MoveResponse, error) {
	pyraMovesMade.Inc()
	me := decoded.You.Body
	var pickDir string

	grid, pf := makePathfinder(decoded)
	target := p.selectTarget(ctx, decoded, pf)

	path, _ := pf.FindPath(me[0].X, me[0].Y, target.Line.B.X, target.Line.B.Y)
	pickDir = me[0].Dir(api.Coord{
		X: path[1].X,
		Y: path[1].Y,
	})

	data, err := json.Marshal(map[string]interface{}{
		"input":    decoded,
		"grid":     grid,
		"target":   target,
		"path":     path,
		"pick_dir": pickDir,
	})
	if err == nil {
		_, err = p.Redis.XAdd(&redis.XAddArgs{
			Stream: "pyra:" + decoded.Game.ID,
			Values: map[string]interface{}{
				"state":    base64.StdEncoding.EncodeToString(data),
				"turn":     decoded.Turn,
				"pick_dir": pickDir,
			},
		}).Result()
		if err != nil {
			ln.Error(ctx, err)
		}
	}

	return &api.MoveResponse{
		Move: pickDir,
	}, nil
}

// End ends a game.
func (Pyra) End(ctx context.Context, sr api.SnakeRequest) error {
	pyraGamesEnded.Inc()
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
