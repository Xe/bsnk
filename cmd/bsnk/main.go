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
	"github.com/go-redis/redis"
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

func middlewareMetrics(family string, next http.Handler) http.Handler {
	cnt := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: family+"_handler_requests_total",
			Help: "Total number of request/responses by HTTP status code.",
		},
		[]string{"code"},
	)
	cnt.WithLabelValues("200")
	cnt.WithLabelValues("500")
	cnt.WithLabelValues("503")

	if err := prometheus.Register(cnt); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			cnt = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			panic(err)
		}
	}

	hst := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: family+"_handler_duration",
		Help: family+" handler request duration.",
	}, []string{"handler", "method"})

	if err := prometheus.Register(hst); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			cnt = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			panic(err)
		}
	}

	gge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: family+"_handler_requests_in_flight",
		Help: "Current number of requests being served.",
	})
	if err := prometheus.Register(gge); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			gge = are.ExistingCollector.(prometheus.Gauge)
		} else {
			panic(err)
		}
	}

	return promhttp.InstrumentHandlerDuration(hst, promhttp.InstrumentHandlerCounter(cnt, promhttp.InstrumentHandlerInFlight(gge, next)))
}

var (
	port          = flag.String("port", "5000", "http port to listen on")
	gitRev        = flag.String("git-rev", "", "if set, use this git revision for the color code")
	pyraMinLength = flag.Int("pyra-min-length", 8, "min length for pyra")
	redisURL      = flag.String("redis-url", "", "redis URL")
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

	options, err := redis.ParseURL(*redisURL)
	if err != nil {
		ln.FatalErr(ctx, err, ln.F{"redis_url": *redisURL})
	}
	c := redis.NewClient(options)

	c.WrapProcess(func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		hst := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "redis_command_duration",
			Help: "Redis command duration",
		}, []string{"verb"})

		return func(cmd redis.Cmder) error {
			t0 := time.Now()
			err := old(cmd)
			hst.With(prometheus.Labels{"verb": cmd.Name()}).Observe(float64(time.Since(t0)))
			return err
		}
	})

	http.HandleFunc("/", index)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", health)
	http.Handle("/garen/", middlewareMetrics("garen", middlewareSpan("garen", api.Server{
		Brain: snakes.Garen{},
		Name:  "garen",
	})))
	http.Handle("/greedy/", middlewareMetrics("greedy", middlewareSpan("greedy", api.Server{
		Brain: snakes.Greedy{
			Redis: c,
		},
		Name: "greedy",
	})))
	http.Handle("/erratic/", middlewareMetrics("erratic", middlewareSpan("erratic", api.Server{
		Brain: snakes.Erratic{},
		Name:  "erratic",
	})))
	http.Handle("/pyra/", middlewareMetrics("pyra", middlewareSpan("pyra", api.Server{
		Brain: snakes.Pyra{
			Redis:     c,
			MinLength: *pyraMinLength,
		},
		Name: "pyra",
	})))

	ln.Log(ctx, ln.Info("booting"))
	ln.FatalErr(ctx, http.ListenAndServe(
		":"+*port,
		middlewareGitRev(ex.HTTPLog(http.DefaultServeMux)),
	), ln.F{
		"git_rev": *gitRev,
		"port":    *port,
	})
}
