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
	res.Header().Set("Content-Type", "text/html")
	res.Write([]byte("<p>Battlesnake documentation can be found at <a href=\"https://docs.battlesnake.io\">https://docs.battlesnake.io</a>.</p>"))
}

func health(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "OK", http.StatusOK)
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

func middlewareGitRev(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Git-Rev", *gitRev)
		next.ServeHTTP(w, r)
	})
}

var (
	port          = flag.String("port", "5000", "http port to listen on")
	gitRev        = flag.String("git-rev", "", "if set, use this git revision for the color code")
	pyraMinLength = flag.Int("pyra-min-length", 8, "min length for pyra")
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
	http.HandleFunc("/health", health)
	http.Handle("/garen/", middlewareSpan("garen", api.Server{
		Brain: snakes.Garen{},
		Name:  "garen",
	}))
	http.Handle("/greedy/", middlewareSpan("greedy", api.Server{
		Brain: snakes.Greedy{},
		Name:  "greedy",
	}))
	http.Handle("/erratic/", middlewareSpan("erratic", api.Server{
		Brain: snakes.Erratic{},
		Name:  "erratic",
	}))
	http.Handle("/pyra/", middlewareSpan("pyra", api.Server{
		Brain: snakes.Pyra{
			MinLength: *pyraMinLength,
		},
		Name: "pyra",
	}))

	ln.Log(ctx, ln.Info("booting"))
	ln.FatalErr(ctx, http.ListenAndServe(
		":"+*port,
		middlewareGitRev(ex.HTTPLog(http.DefaultServeMux)),
	), ln.F{
		"git_rev": *gitRev,
		"port":    *port,
	})
}
