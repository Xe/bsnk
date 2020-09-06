package main

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Xe/bsnk/api"
	"github.com/Xe/bsnk/snakes"
	"github.com/facebookgo/flagenv"
	"github.com/povilasv/prommod"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "handler_requests_total",
			Help: "Total number of request/responses by HTTP status code.",
		}, []string{"handler", "code"})

	requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "handler_request_duration",
		Help: "Handler request duration.",
	}, []string{"handler", "method"})

	requestInFlight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "handler_requests_in_flight",
		Help: "Current number of requests being served.",
	}, []string{"handler"})
)

func init() {
	prometheus.Register(requestCounter)
	prometheus.Register(requestDuration)
	prometheus.Register(requestInFlight)
}

func middlewareMetrics(family string, next http.Handler) http.Handler {
	return promhttp.InstrumentHandlerDuration(
		requestDuration.MustCurryWith(prometheus.Labels{"handler": family}),
		promhttp.InstrumentHandlerCounter(requestCounter.MustCurryWith(prometheus.Labels{"handler": family}),
			promhttp.InstrumentHandlerInFlight(requestInFlight.With(prometheus.Labels{"handler": family}), next),
		),
	)
}

var (
	port          = flag.String("port", "5000", "http port to listen on")
	gitRev        = flag.String("git-rev", "", "if set, use this git revision for the color code")
	pyraMinLength = flag.Int("pyra-min-length", 8, "min length for pyra")
)

func createSnake(name string, ai api.AI) http.Handler {
	return middlewareMetrics(name,
		middlewareSpan(name, api.Server{
			Brain: ai,
			Name:  name,
		}),
	)
}

func init() {
	ln.AddFilter(ex.NewGoTraceLogger())

	trace.AuthRequest = func(_ *http.Request) (bool, bool) {
		return true, true
	}
}

func vars(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	e.Encode(map[string]interface{}{
		"git_rev":         *gitRev,
		"pyra_min_length": *pyraMinLength,
	})
}

func main() {
	flagenv.Parse()
	flag.Parse()

	ctx := opname.With(context.Background(), "main")

	prometheus.Register(prommod.NewCollector("bsnk"))

	http.HandleFunc("/", index)
	http.HandleFunc("/vars", vars)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", health)
	http.Handle("/garen/", createSnake("garen", snakes.Garen{}))
	http.Handle("/greedy/", createSnake("greedy", &snakes.Greedy{}))
	http.Handle("/erratic/", createSnake("erratic", snakes.Erratic{}))
	http.Handle("/pyra/", createSnake("pyra", &snakes.Pyra{
		MinLength: *pyraMinLength,
	}))
	http.Handle("/sunset/", createSnake("sunset", snakes.Sunset{}))

	ln.Log(ctx, ln.Info("booting"))
	ln.FatalErr(ctx, http.ListenAndServe(
		":"+*port,
		middlewareGitRev(ex.HTTPLog(http.DefaultServeMux)),
	), ln.F{
		"git_rev": *gitRev,
		"port":    *port,
	})
}
