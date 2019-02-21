package api

import (
	"encoding/json"
	"net/http"
)

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (l Coord) Dir (r Coord) string {
	switch{
	case l.X >= r.X:
		return "right"
	case l.X < r.X:
		return "left"
	case l.Y >= r.Y:
		return "up"
	case l.Y < r.X:
		return "down"
	}

	return "how"
}

func (l Coord) Eq (r Coord) bool {
	return l.X == r.X && l.Y == r.Y
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
