// +build ignore

package main

import (
	"github.com/kr/pretty"
	as "github.com/prettymuchbryce/goeasystar"
)

func main() {
	pathfinder := as.NewPathfinder()

	var grid [][]int
	grid = append(grid, []int{0, 0, 0, 0, 0})
	grid = append(grid, []int{0, 0, 0, 0, 0})
	grid = append(grid, []int{0, 0, 0, 0, 0})
	grid = append(grid, []int{0, 0, 0, 0, 0})
	grid = append(grid, []int{0, 0, 0, 0, 0})

	pathfinder.SetGrid(grid)
	pathfinder.DisableDiagonals()
	pathfinder.SetAcceptableTiles([]int{0})

	path, err := pathfinder.FindPath(4, 4, 2, 2)
	pretty.Println(path)
	pretty.Println(err)
}
