package api

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"within.website/ln"
	"within.website/ln/opname"
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

// AI is an individual snake AI.
type AI interface {
	Ping() (*PingResponse, error)
	Start(ctx context.Context, sr SnakeRequest) error
	Move(ctx context.Context, sr SnakeRequest) (*MoveResponse, error)
	End(ctx context.Context, sr SnakeRequest) error
}

// Server wraps an AI.
type Server struct {
	Brain AI
	Name  string
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Snake-AI", s.Name)

	if r.Method != http.MethodPost {
		http.Error(w, s.Name+" OK", http.StatusOK)
		return
	}

	var result ln.Fer
	var err error

	decoded := SnakeRequest{}
	err = DecodeSnakeRequest(r, &decoded)
	if err != nil {
		ln.Error(r.Context(), err)
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	ctx := ln.WithF(r.Context(), decoded.F())
	ctx = opname.With(ctx, s.Name)

	switch filepath.Base(r.URL.Path) {
	case "start":
		ctx := opname.With(ctx, "start-game")
		err = s.Brain.Start(ctx, decoded)
		gamesStarted.With(prometheus.Labels{"brain": s.Name}).Inc()
		if err == nil {
			ln.Log(ctx, decoded, result)
		}
		result = ln.F{}
	case "move":
		ctx := opname.With(ctx, "move")
		result, err = s.Brain.Move(ctx, decoded)
		movesMade.With(prometheus.Labels{"brain": s.Name}).Inc()
		if err == nil {
			ln.Log(ctx, decoded, result)
		}
	case "end":
		ctx := opname.With(ctx, "end")
		err = s.Brain.End(ctx, decoded)
		gamesEnded.With(prometheus.Labels{"brain": s.Name}).Inc()
		ln.Log(ctx, decoded)
	case "":
		result, err = s.Brain.Ping()
	}

	if err != nil {
		ln.Error(ctx, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
	w.Write([]byte("\n"))
}
