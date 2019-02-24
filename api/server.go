package api

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

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
	if r.Method != http.MethodPost {
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

	ctx := ln.WithF(r.Context(), decoded.F())

	switch filepath.Base(r.URL.Path) {
	case "start":
		ctx := opname.With(ctx, "start-game")
		result, err = s.Brain.Start(ctx, decoded)
		ln.Log(ctx, decoded, result)
	case "move":
		ctx := opname.With(ctx, "move")
		result, err = s.Brain.Move(ctx, decoded)
		ln.Log(ctx, decoded, result)
	case "end":
		ctx := opname.With(ctx, "end")
		err = s.Brain.End(ctx, decoded)
		ln.Log(ctx, decoded)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
	w.Write([]byte("\n"))
}
