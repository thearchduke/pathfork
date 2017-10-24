package main

import (
	"flag"
	"net/http"
	"os"

	"bitbucket.org/jtyburke/pathfork/app"
	"bitbucket.org/jtyburke/pathfork/app/config"
	"github.com/golang/glog"
	"github.com/gorilla/context"
	_ "github.com/lib/pq"
)

const staticDir = "static"

func determineListenAddress() string {
	port := os.Getenv("PORT")
	if port == "" {
		glog.Info("No $PORT env variable found, defaulting to 8080")
		return ":8080"
	}
	return ":" + port
}

func main() {
	flag.Parse()
	tr, db, store := pathfork.InitApp()
	defer db.DB.Close()
	glog.Info("Starting static server")
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(config.StaticPath)))
	http.Handle(pathfork.StaticRoute, fs)
	for _, route := range pathfork.FrontEndRoutes {
		http.HandleFunc(route.Path, pathfork.WrapFrontEndHandler(route.Handler, tr, db, store))
	}
	port := determineListenAddress()
	glog.Infof("Serving Pathfork on port %v", port)
	http.ListenAndServe(port, context.ClearHandler(http.DefaultServeMux))
	glog.Flush()
}
