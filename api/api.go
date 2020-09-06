package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"within.website/ln"
)

// Coord is an X,Y coordinate pair.
type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (l Coord) String() string {
	return fmt.Sprintf("(%d,%d)", l.X, l.Y)
}

// Dir computes the net immediate direction from point l to point r
func (l Coord) Dir(r Coord) string {
	switch {
	case l.X < r.X:
		return "right"
	case l.X > r.X:
		return "left"
	case l.Y < r.Y:
		return "up"
	case l.Y > r.Y:
		return "down"
	}

	return "how"
}

// Eq checks if one Coord equals another.
func (l Coord) Eq(r Coord) bool {
	return l.X == r.X && l.Y == r.Y
}

func (l Coord) Left() Coord {
	return Coord{
		X: l.X - 1,
		Y: l.Y,
	}
}

func (l Coord) Right() Coord {
	return Coord{
		X: l.X + 1,
		Y: l.Y,
	}
}

func (l Coord) Up() Coord {
	return Coord{
		X: l.X,
		Y: l.Y + 1,
	}
}

func (l Coord) Down() Coord {
	return Coord{
		X: l.X,
		Y: l.Y - 1,
	}
}

// Snake is a competitor.
type Snake struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Health int     `json:"health"`
	Body   []Coord `json:"body"`
}

// Board is the game board.
type Board struct {
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Food   []Coord `json:"food"`
	Snakes []Snake `json:"snakes"`
}

// Inside checks if a point is inside the board.
func (b Board) Inside(x Coord) bool {
	switch {
	case x.X >= b.Width:
		return false
	case x.Y >= b.Height:
		return false
	}

	if x.X < 0 {
		return false
	}

	if x.Y < 0 {
		return false
	}

	return true
}

// DeadlyAdjacent returns all of the adjacent deadly coordinates from this one.
func (b Board) DeadlyAdjacent(x Coord) []Coord {
	var result []Coord
	for _, place := range []Coord{
		x.Up(),
		x.Left(),
		x.Down(),
		x.Right(),
	} {
		if b.IsDeadly(place) {
			result = append(result, place)
		}
	}

	return result
}

// IsDeadly checks if a point would kill a snake if it moved into it.
func (b Board) IsDeadly(x Coord) bool {
	if !b.Inside(x) {
		return true
	}

	for _, sn := range b.Snakes {
		for _, bd := range sn.Body {
			if x.Eq(bd) {
				return true
			}
		}
	}

	return false
}

type Game struct {
	ID string `json:"id"`
}

type SnakeRequest struct {
	Game  Game  `json:"game"`
	Turn  int   `json:"turn"`
	Board Board `json:"board"`
	You   Snake `json:"you"`
}

func (sr SnakeRequest) F() ln.F {
	return ln.F{
		"game_id":      sr.Game.ID,
		"turn":         sr.Turn,
		"food_count":   len(sr.Board.Food),
		"snakes_count": len(sr.Board.Snakes),
		"my_health":    sr.You.Health,
	}
}

type PingResponse struct {
	APIVersion string `json:"apiversion,omitempty"`
	Color      string `json:"color,omitempty"`
	HeadType   string `json:"head,omitempty"`
	TailType   string `json:"tail,omitempty"`
}

func (s PingResponse) F() ln.F {
	return ln.F{
		"response_color": s.Color,
	}
}

type MoveResponse struct {
	Move string `json:"move"`
}

func (m MoveResponse) F() ln.F {
	return ln.F{
		"response_move": m.Move,
	}
}

func DecodeSnakeRequest(req *http.Request, decoded *SnakeRequest) error {
	err := json.NewDecoder(req.Body).Decode(&decoded)
	return err
}
