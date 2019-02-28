package snakes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/Xe/bsnk/api"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
	"within.website/ln"
)

// Greedy is a greedy snake AI. It will try to get as long as it can as fast as
// it can.
type Greedy struct {
	Redis *redis.Client
}

// Start starts a game.
func (Greedy) Start(ctx context.Context, gs api.SnakeRequest) (*api.StartResponse, error) {
	gamesStarted.With(prometheus.Labels{"brain": "greedy"}).Inc()
	return &api.StartResponse{
		Color: "#c79dd7",
	}, nil
}

// Move responds with the snake's movements for a given Turn.
func (g Greedy) Move(ctx context.Context, decoded api.SnakeRequest) (*api.MoveResponse, error) {
	movesMade.With(prometheus.Labels{"brain": "greedy"}).Inc()
	me := decoded.You.Body
	var pickDir string

	grid, pf := makePathfinder(decoded)
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

	data, err := json.Marshal(map[string]interface{}{
		"input":    decoded,
		"grid":     grid,
		"target":   target,
		"path":     path,
		"pick_dir": pickDir,
	})
	if err == nil {
		_, err = g.Redis.XAdd(&redis.XAddArgs{
			Stream: "greedy:" + decoded.Game.ID,
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
func (Greedy) End(ctx context.Context, sr api.SnakeRequest) error {
	gamesEnded.With(prometheus.Labels{"brain": "greedy"}).Inc()
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

func selectGreedy(gs api.SnakeRequest) api.Coord {
	me := gs.You.Body
	var target api.Coord
	var foundTarget bool
	var distance float64 = 99999999999

	for _, fd := range gs.Board.Food {
		if sc := manhattan(me[0], fd); sc < distance && !gs.Board.IsDeadly(fd) {
			for _, place := range []api.Coord{fd.Up(), fd.Down(), fd.Left(), fd.Right()} {
				if gs.Board.IsDeadly(place) {
					goto nextFood
				}
			}

			distance = sc
			target = fd
			foundTarget = true
		}
	nextFood:
	}

	if !foundTarget {
		tail := me[len(me)-1]
		for _, place := range []api.Coord{tail.Up(), tail.Down(), tail.Left(), tail.Right()} {
			if !gs.Board.IsDeadly(place) {
				target = place
			}
		}
	}

	return target
}

func logCoords(pfx string, coord api.Coord) ln.F {
	return ln.F{
		pfx + "_x,y": fmt.Sprintf("(%d,%d)", coord.X, coord.Y),
	}
}
