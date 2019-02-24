package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/Xe/bsnk/api"
	"github.com/facebookgo/flagenv"
	"github.com/go-redis/redis"
	"github.com/prettymuchbryce/goeasystar"
	"golang.org/x/net/trace"
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
	port          = flag.String("port", "5000", "http port to listen on")
	color         = flag.String("color", "#c79dd7", "snake color code to use")
	gitRev        = flag.String("git-rev", "", "if set, use this git revision for the color code")
	redisURL      = flag.String("redis-url", "", "URL for redis storage of battlesnake info")
	tracingFamily = flag.String("tracing-family", "sparklebutt", "tracing family to use")
)

func init() {
	ln.AddFilter(ex.NewGoTraceLogger())

	trace.AuthRequest = func(_ *http.Request) (bool, bool) {
		return true, true
	}
}

func middlewareSpan(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sp := trace.New(*tracingFamily, r.URL.Path)
		defer sp.Finish()
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		ctx = trace.NewContext(ctx, sp)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	flagenv.Parse()
	flag.Parse()

	ctx := opname.With(context.Background(), "main")

	if *redisURL == "" {
		ln.Fatal(ctx, ln.Info("-redis-url not defined and is needed"))
	}

	opt, err := redis.ParseURL(*redisURL)
	if err != nil {
		ln.FatalErr(ctx, err)
	}

	b := bot{
		rc:          redis.NewClient(opt),
		gameCounter: expvar.NewInt("games_begun"),
		endCounter:  expvar.NewInt("games_ended"),
		moveCounter: expvar.NewInt("moves_made"),
		pingCounter: expvar.NewInt("ping_counter"),
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/start", b.start)
	http.HandleFunc("/move", b.move)
	http.HandleFunc("/end", b.end)
	http.HandleFunc("/ping", b.ping)

	f := ln.F{
		"port": *port,
	}

	if *color != "" {
		f["color"] = *color
	}

	if *gitRev != "" {
		f["git-rev"] = *gitRev
	}

	ln.Log(ctx, f, ln.Info("booting"))
	ln.FatalErr(ctx, http.ListenAndServe(":"+*port, middlewareSpan(ex.HTTPLog(http.DefaultServeMux))), f)
}

type bot struct {
	rc          *redis.Client
	gameCounter *expvar.Int
	moveCounter *expvar.Int
	endCounter  *expvar.Int
	pingCounter *expvar.Int
}

func (b bot) start(res http.ResponseWriter, req *http.Request) {
	decoded := api.SnakeRequest{}
	err := api.DecodeSnakeRequest(req, &decoded)
	if err != nil {
		log.Printf("Bad start request: %v", err)
		http.Error(res, "bad json", http.StatusBadRequest)
		return
	}

	f := ln.F{
		"game_id":   decoded.Game.ID,
		"turn":      decoded.Turn,
		"board_y":   decoded.Board.Height,
		"board_x":   decoded.Board.Width,
		"my_health": decoded.You.Health,
	}

	b.gameCounter.Add(1)

	ctx := opname.With(req.Context(), "game-start")

	clr := *color

	if *gitRev != "" {
		rev := *gitRev
		clr = "#" + rev[0:6]
	}

	respond(res, api.StartResponse{
		Color: clr,
	})

	data, err := json.Marshal(decoded)
	if err != nil {
		// should not happen
		panic(err)
	}

	rc := b.rc.WithContext(ctx)

	id, err := rc.XAdd(&redis.XAddArgs{
		Stream: "game:" + decoded.Game.ID,
		Values: map[string]interface{}{
			"turn":  decoded.Turn,
			"data":  base64.StdEncoding.EncodeToString(data),
			"color": clr,
		},
	}).Result()
	if err != nil {
		ln.Error(ctx, err, ln.Info("can't add to stream"))
	} else {
		f["stream_id"] = id
	}

	ln.Log(ctx, f, ln.Info("starting game"))
}

func manhattan(l, r api.Coord) float64 {
	absX := r.X - l.X
	if absX < 0 {
		absX = -absX
	}

	absY := r.Y - l.Y
	if absY < 0 {
		absY = -absY
	}

	return float64(absX + absY)
}

func selectFood(gs api.SnakeRequest) api.Coord {
	me := gs.You.Body
	var target api.Coord
	var foundTarget bool
	var distance float64 = 99999999999

	for _, fd := range gs.Board.Food {
		if sc := manhattan(me[0], fd); sc < distance {
			distance = sc
			target = fd
			foundTarget = true
		}
	}

	if !foundTarget {
		target = api.Coord{
			X: rand.Intn(gs.Board.Width),
			Y: rand.Intn(gs.Board.Height),
		}
	}

	return target
}

func (b bot) move(res http.ResponseWriter, req *http.Request) {
	ctx := opname.With(req.Context(), "move")
	rc := b.rc.WithContext(ctx)
	decoded := api.SnakeRequest{}
	err := api.DecodeSnakeRequest(req, &decoded)
	if err != nil {
		log.Printf("Bad move request: %v", err)
		http.Error(res, "bad json", http.StatusBadRequest)
		return
	}

	b.moveCounter.Add(1)

	me := decoded.You.Body
	var pickDir string
	var target api.Coord

	switch {
	default:
		target = selectFood(decoded)
	}

	pf := goeasystar.NewPathfinder()
	pf.DisableCornerCutting()
	pf.DisableDiagonals()
	pf.SetAcceptableTiles([]int{1})

	var grid [][]int
	grid = make([][]int, decoded.Board.Height)
	for i := range grid {
		grid[i] = make([]int, decoded.Board.Width)
		for j := range grid[i] {
			grid[i][j] = 1
		}
	}

	pf.SetGrid(grid)

	for _, sk := range decoded.Board.Snakes {
		for _, pt := range sk.Body {
			pf.AvoidAdditionalPoint(pt.X, pt.Y)

			lf := pt.Left()
			pf.SetAdditionalPointCost(lf.X, lf.Y, 300)
			lf = pt.Right()
			pf.SetAdditionalPointCost(lf.X, lf.Y, 300)
			lf = pt.Up()
			pf.SetAdditionalPointCost(lf.X, lf.Y, 300)
			lf = pt.Down()
			pf.SetAdditionalPointCost(lf.X, lf.Y, 300)
		}
	}

	path, err := pf.FindPath(me[0].X, me[0].Y, target.X, target.Y)
	if err != nil {
		ln.Error(ctx, err)
		target = api.Coord{
			X: rand.Intn(decoded.Board.Width),
			Y: rand.Intn(decoded.Board.Height),
		}

		path, err = pf.FindPath(me[0].X, me[0].Y, target.X, target.Y)
	}
	if len(path) != 0 {
		pickDir = me[0].Dir(api.Coord{
			X: path[1].X,
			Y: path[1].Y,
		})
	}

	f := ln.F{
		"game_id":   decoded.Game.ID,
		"turn":      decoded.Turn,
		"board_y":   decoded.Board.Height,
		"board_x":   decoded.Board.Width,
		"my_health": decoded.You.Health,
		"picking":   pickDir,
		"distance":  manhattan(me[0], target),
	}
	f.Extend(logCoords("my_head", decoded.You.Body[0]))
	f.Extend(logCoords("target", target))

	respond(res, api.MoveResponse{
		Move: pickDir,
	})

	data, err := json.Marshal(decoded)
	if err != nil {
		// should not happen
		panic(err)
	}

	id, err := rc.XAdd(&redis.XAddArgs{
		Stream: "game:" + decoded.Game.ID,
		Values: map[string]interface{}{
			"turn":   decoded.Turn,
			"data":   base64.StdEncoding.EncodeToString(data),
			"me":     fmt.Sprintf("(%d,%d)", me[0].X, me[0].Y),
			"target": fmt.Sprintf("(%d,%d)", target.X, target.Y),
			"picked": pickDir,
		},
	}).Result()
	if err != nil {
		ln.Error(ctx, err, ln.Info("can't add to stream"))
	} else {
		f["stream_id"] = id
	}

	ln.Log(ctx, f, ln.Info("moving"))
}

func logCoords(pfx string, coord api.Coord) ln.F {
	return ln.F{
		pfx + "_x,y": fmt.Sprintf("(%d,%d)", coord.X, coord.Y),
	}
}

func (b bot) end(res http.ResponseWriter, req *http.Request) {
	b.endCounter.Add(1)
	return
}

func (b bot) ping(res http.ResponseWriter, req *http.Request) {
	b.pingCounter.Add(1)
	return
}
