package main

import (
	"context"
	"net/http"
	"os"

	"within.website/ln"
	"within.website/ln/ex"
	"within.website/ln/opname"
)

func index(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Battlesnake documentation can be found at <a href=\"https://docs.battlesnake.io\">https://docs.battlesnake.io</a>."))
}

var port = os.Getenv("PORT")

func main() {
	http.HandleFunc("/", index)

	if port == "" {
		port =
			"5000"
	}

	f := ln.F{
		"port": port,
	}

	ctx := opname.With(context.Background(), "main")
	ln.Log(ctx, f, ln.Info("booting"))
	ln.FatalErr(ctx, http.ListenAndServe(":"+port, ex.HTTPLog(http.DefaultServeMux)), f)
}
