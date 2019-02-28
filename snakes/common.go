package snakes

import (
	"github.com/Xe/bsnk/api"
	"github.com/prettymuchbryce/goeasystar"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	gamesStarted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "games_started",
		Help: "The number of games started",
	}, []string{"brain"})

	movesMade = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "game_moves_made",
		Help: "The number of moves made",
	}, []string{"brain"})

	gamesEnded = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "games_ended",
		Help: "The number of games ended",
	}, []string{"brain"})
)

func makePathfinder(decoded api.SnakeRequest) ([][]int, *goeasystar.Pathfinder) {
	pf := goeasystar.NewPathfinder()
	pf.DisableCornerCutting()
	pf.DisableDiagonals()
	pf.SetAcceptableTiles([]int{1, 2, 5, 8})

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
		var headDir string
		var theirNext api.Coord
		if len(sk.Body) < 2 {
			goto skipHead
		}
		headDir = sk.Body[1].Dir(sk.Body[0])
		{
			switch headDir {
			case "left":
				theirNext = sk.Body[0].Left()
			case "right":
				theirNext = sk.Body[0].Right()
			case "up":
				theirNext = sk.Body[0].Up()
			case "down":
				theirNext = sk.Body[0].Down()
			default:
				goto skipHead
			}
			if decoded.Board.Inside(theirNext) {
				pf.AvoidAdditionalPoint(theirNext.X, theirNext.Y)
			}
		}
	skipHead:

		for _, pt := range sk.Body {
			pf.AvoidAdditionalPoint(pt.X, pt.Y)
			grid[pt.X][pt.Y] = 9

			if sk.ID != decoded.You.ID {
				for _, st := range []api.Coord{
					pt.Up(),
					pt.Up().Up(),
					pt.Left(),
					pt.Left().Left(),
					pt.Right(),
					pt.Right().Right(),
					pt.Down(),
					pt.Down().Down(),
				} {
					if decoded.Board.Inside(st) {
						grid[st.X][st.Y] = 5
						pf.SetAdditionalPointCost(st.X, st.Y, 5)
					}
				}
			}
		}
	}

	return grid, pf
}
