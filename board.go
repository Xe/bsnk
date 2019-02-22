package main

import (
	"log"

	"github.com/Xe/bsnk/api"
)

type Board struct {
	api.Board
	Self api.Snake
}

func MakeBoard(sr *api.SnakeRequest) *Board {
	return &Board{
		Board: sr.Board,
		Self:  sr.You,
	}
}

func (b *Board) GetFoods() []Cell {
	var result []Cell

	for _, fd := range b.Food {
		result = append(result, *b.makeCell(fd.X, fd.Y))
	}

	return result
}

func (b *Board) HeadCoords() api.Coord {
	return b.Self.Body[0]
}
func (b *Board) GetSelfHead() *Cell {
	return b.makeCell(b.Self.Body[0].X, b.Self.Body[0].Y)
}

func (b *Board) makeCell(x, y int) *Cell {
	log.Printf("making cell at %d,%d", x, y)
	c := api.Coord{
		X: x,
		Y: y,
	}

	result := &Cell{
		ref: b,
	}

	if !b.isInBoard(c) {
		result.Contents = Wall
		return result
	}

	for _, snk := range b.Snakes {
		for i, seg := range snk.Body {
			if c.Eq(seg) {
				switch i {
				case 0:
					result.Contents = EnemySnakeHead
				default:
					result.Contents = EnemySnake
				}
				return result
			}
		}
	}

	for _, myBody := range b.Self.Body {
		if c.Eq(myBody) {
			result.Contents = MySnake
			return result
		}
	}

	for _, food := range b.Food {
		if c.Eq(food) {
			result.Contents = Food
			return result
		}
	}

	return result
}

func (b Board) isInBoard(inp api.Coord) bool {
	if inp.X >= b.Width {
		return false
	}

	if inp.Y >= b.Height {
		return false
	}

	if inp.X < 0 {
		return false
	}

	if inp.Y < 0 {
		return false
	}

	return true
}

type CellContents int

const (
	None CellContents = iota
	MySnake
	EnemySnake
	EnemySnakeHead
	Food
	Wall
)

type Cell struct {
	ref      *Board
	Coord    api.Coord
	Contents CellContents
}

func (c Cell) neighbor(relX, relY int) api.Coord {
	return api.Coord{
		X: c.Coord.X+relX,
		Y: c.Coord.Y+relY,
	}
}

func (c Cell) up() api.Coord {
	return c.neighbor(0, 1)
}

func (c Cell) down() api.Coord {
	return c.neighbor(0, -1)
}

func (c Cell) left() api.Coord {
	return c.neighbor(-1, 0)
}

func (c Cell) right() api.Coord {
	return c.neighbor(1, 0)
}

func (c *Cell) PathNeighbors() []api.Coord {
	var result []api.Coord
	for _, side := range []api.Coord{c.up(), c.down(), c.left(), c.right()} {
		if pathNeighborCost(c.ref.makeCell(side.X, side.Y)) != doNotMove {
			result = append(result, side)
		}
	}

	return result
}

// pathfinding cost hacking
const (
	doNotMove = 99999
	getThis   = 500
	normal    = 100
)

func pathNeighborCost(to *Cell) float64 {
	switch to.Contents {
	case Food:
		return getThis
	case None:
		return normal
	}

	return doNotMove
}

func (c Cell) PathEstimatedCost(to Cell) float64 {
	absX := to.Coord.X - c.Coord.X
	if absX < 0 {
		absX = -absX
	}

	absY := to.Coord.Y - c.Coord.Y
	if absY < 0 {
		absY = -absY
	}

	return float64(absX + absY)
}
