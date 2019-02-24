package api

import (
	"encoding/json"
	"net/http"
)

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (l Coord) Dir(r Coord) string {
	switch {
	case l.X > r.X:
		return "right"
	case l.X < r.X:
		return "left"
	case l.Y > r.Y:
		return "up"
	case l.Y < r.X:
		return "down"
	}

	return "how"
}

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

type Snake struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Health int     `json:"health"`
	Body   []Coord `json:"body"`
}

type Board struct {
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Food   []Coord `json:"food"`
	Snakes []Snake `json:"snakes"`
}

func (b Board) IsDeadly(x Coord) bool {
	switch {
	case x.X >= b.Width:
		return true
	case x.Y >= b.Height:
		return true
	}

	for _, sn := range b.Snakes {
		for _, bd := range sn.Body {
			if bd.Eq(x) {
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

type StartResponse struct {
	Color string `json:"color,omitempty"`
}

type MoveResponse struct {
	Move string `json:"move"`
}

func DecodeSnakeRequest(req *http.Request, decoded *SnakeRequest) error {
	err := json.NewDecoder(req.Body).Decode(&decoded)
	return err
}
