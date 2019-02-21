package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Xe/bsnk/api"
	"github.com/facebookgo/flagenv"
	"github.com/kr/pretty"
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

	if me.left().Coord.X < me.Coord.X {
		pretty.Println(decoded)
		panic(nil)
	}

	var target Cell
	var targetCost float64 = 99999
	var goalStr = "nothing"
	var goal Cell

	for _, fd := range b.GetFoods() {
		f := logCoords("food", fd.Coord)

		distance := me.PathEstimatedCost(fd)
		f["distance"] = distance
		ln.Log(ctx, ln.Info("found distance to food"), f)

		if distance < targetCost {
			for _, side := range []Cell{me.up(), me.down(), me.left(), me.right()} {
				ln.Log(ctx, ln.Info("comparing side"), logCoords("at", side.Coord), f)
				if side.PathEstimatedCost(fd) < distance {
					target = side
					targetCost = distance
					goalStr = "food"
					goal = fd
				}
			}
		}
	}

	pickDir = me.Coord.Dir(target.Coord)

	ctx := opname.With(req.Context(), "make-move")
	ln.Log(ctx,
		ln.F{
			"game_id":   decoded.Game.ID,
			"turn":      decoded.Turn,
			"board_y":   decoded.Board.Height,
			"board_x":   decoded.Board.Width,
			"my_health": decoded.You.Health,
			"my_head_x": decoded.You.Body[0].X,
			"my_head_y": decoded.You.Body[0].Y,
			"picking":   pickDir,
			"goal":      goalStr,
		},
		logCoords("goal", goal.Coord),
		logCoords("left", me.left().Coord),
		logCoords("right", me.right().Coord),
		logCoords("up", me.up().Coord),
		logCoords("down", me.down().Coord),
	)

	respond(res, api.MoveResponse{
		Move: pickDir,
	})
}

func logCoords(pfx string, coord api.Coord) ln.F {
	return ln.F{
		pfx + "_x,y": fmt.Sprintf("(%d,%d)", coord.X, coord.Y),
	}
}

func End(res http.ResponseWriter, req *http.Request) {
	return
}

func Ping(res http.ResponseWriter, req *http.Request) {
	return
}
