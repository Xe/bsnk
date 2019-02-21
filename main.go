package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/Xe/bsnk/api"
	"github.com/beefsack/go-astar"
	"github.com/facebookgo/flagenv"
	"within.website/ln"
	"within.website/ln/ex"
	"within.website/ln/opname"
)

func respond(res http.ResponseWriter, obj interface{}) {
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(obj)
	res.Write([]byte("\n"))
}

func index(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Battlesnake documentation can be found at <a href=\"https://docs.battlesnake.io\">https://docs.battlesnake.io</a>."))
}

var (
	port  = flag.String("port", "5000", "http port to listen on")
	color = flag.String("color", "#c79dd7", "snake color code to use")
)

func main() {
	flagenv.Parse()
	flag.Parse()

	http.HandleFunc("/", index)
	http.HandleFunc("/start", Start)
	http.HandleFunc("/move", Move)
	http.HandleFunc("/end", End)
	http.HandleFunc("/ping", Ping)

	f := ln.F{
		"port": *port,
	}

	ctx := opname.With(context.Background(), "main")
	ln.Log(ctx, f, ln.Info("booting"))
	ln.FatalErr(ctx, http.ListenAndServe(":"+*port, ex.HTTPLog(http.DefaultServeMux)), f)
}

func Start(res http.ResponseWriter, req *http.Request) {
	decoded := api.SnakeRequest{}
	err := api.DecodeSnakeRequest(req, &decoded)
	if err != nil {
		log.Printf("Bad start request: %v", err)
	}

	ctx := opname.With(req.Context(), "game-start")
	ln.Log(ctx, ln.F{
		"game_id":   decoded.Game.ID,
		"turn":      decoded.Turn,
		"board_y":   decoded.Board.Height,
		"board_x":   decoded.Board.Width,
		"my_health": decoded.You.Health,
	})

	respond(res, api.StartResponse{
		Color: *color,
	})
}

func Move(res http.ResponseWriter, req *http.Request) {
	decoded := api.SnakeRequest{}
	err := api.DecodeSnakeRequest(req, &decoded)
	if err != nil {
		log.Printf("Bad move request: %v", err)
	}

	var pickDir = "down"

	b := MakeBoard(&decoded)
	me := b.GetSelfHead()
	var target api.Coord
	var targetCost float64

	for _, fd := range b.GetFoods() {
		path, distance, found := astar.Path(me, fd)
		if !found {
			// can't get to this food
			continue
		}

		if distance < targetCost {
			target = path[0].(Cell).Coord
			targetCost = distance
		}
	}

	pickDir = me.Coord.Dir(target)

	ctx := opname.With(req.Context(), "make-move")
	ln.Log(ctx, ln.F{
		"game_id":   decoded.Game.ID,
		"turn":      decoded.Turn,
		"board_y":   decoded.Board.Height,
		"board_x":   decoded.Board.Width,
		"my_health": decoded.You.Health,
		"my_head_x": decoded.You.Body[0].X,
		"my_head_y": decoded.You.Body[0].Y,
		"picking":   pickDir,
	})

	respond(res, api.MoveResponse{
		Move: pickDir,
	})
}

func End(res http.ResponseWriter, req *http.Request) {
	return
}

func Ping(res http.ResponseWriter, req *http.Request) {
	return
}
