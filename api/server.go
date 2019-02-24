package api

import (
	"context"
	"encoding/json"
	"net/http"

	"within.website/ln"
	"within.website/ln/opname"
)

// AI is an individual snake AI.
type AI interface {
	Start(ctx context.Context, sr SnakeRequest) (*StartResponse, error)
	Move(ctx context.Context, sr SnakeRequest) (*MoveResponse, error)
	End(ctx context.Context, sr SnakeRequest) error
}

// Server wraps an AI.
type Server struct {
	Brain AI
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/ping":
		http.Error(w, "OK", http.StatusOK)
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

	switch r.URL.Path {
	case "/start":
		ctx := opname.With(r.Context(), "start-game")
		result, err = s.Brain.Start(ctx, decoded)
	case "/move":
		ctx := opname.With(r.Context(), "move")
		result, err = s.Brain.Move(ctx, decoded)
	case "/end":
		ctx := opname.With(r.Context(), "end")
		err = s.Brain.End(ctx, decoded)
	}

	ln.Log(r.Context(), decoded, result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
	w.Write([]byte("\n"))
}
