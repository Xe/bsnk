package main

import (
	"context"
	"flag"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Xe/bsnk/api"
	"github.com/Xe/bsnk/snakes"
	"github.com/facebookgo/flagenv"
	"golang.org/x/net/trace"
	"within.website/ln"
	"within.website/ln/ex"
	"within.website/ln/opname"
)

func index(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Battlesnake documentation can be found at <a href=\"https://docs.battlesnake.io\">https://docs.battlesnake.io</a>."))
}

func middlewareSpan(family string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sp := trace.New(family, filepath.Base(r.URL.Path))
		defer sp.Finish()
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		ctx = trace.NewContext(ctx, sp)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var (
	port   = flag.String("port", "5000", "http port to listen on")
	gitRev = flag.String("git-rev", "", "if set, use this git revision for the color code")
)

func init() {
	ln.AddFilter(ex.NewGoTraceLogger())

	trace.AuthRequest = func(_ *http.Request) (bool, bool) {
		return true, true
	}
}

func main() {
	flagenv.Parse()
	flag.Parse()

	ctx := opname.With(context.Background(), "main")

	http.HandleFunc("/", index)
	http.Handle("/garen/", middlewareSpan("garen", api.Server{Brain: snakes.Garen{}}))

	ln.Log(ctx, ln.Info("booting"))
	ln.FatalErr(ctx, http.ListenAndServe(":"+*port, http.DefaultServeMux))
}
