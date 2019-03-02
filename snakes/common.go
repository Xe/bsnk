package snakes

import (
	"github.com/Xe/bsnk/api"
	"github.com/prettymuchbryce/goeasystar"
)

// SpaceContents is the contents of a single space on the map.
type SpaceContents int

// Tile cost constants
const (
	Nothing   = 1
	Risky     = 7
	SnakeBody = 9
)

func makePathfinder(decoded api.SnakeRequest) ([][]int, *goeasystar.Pathfinder) {
	pf := goeasystar.NewPathfinder()
	pf.DisableCornerCutting()
	pf.DisableDiagonals()
	pf.SetAcceptableTiles([]int{Nothing, Risky})

	var grid [][]int
	grid = make([][]int, decoded.Board.Height)
	for i := range grid {
		grid[i] = make([]int, decoded.Board.Width)

		for j := range grid[i] {
			grid[i][j] = Nothing
		}
	}

	for _, sk := range decoded.Board.Snakes {
		for _, pt := range sk.Body {
			grid[pt.X][pt.Y] = SnakeBody

			if sk.ID != decoded.You.ID {
				for _, st := range []api.Coord{
					pt.Up(),
					pt.Left(),
					pt.Right(),
					pt.Down(),
				} {
					if decoded.Board.Inside(st) {
						grid[st.X][st.Y] = Risky
					}
				}
			}
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
		}
	}

	return grid, pf
}
