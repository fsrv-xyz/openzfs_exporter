package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"openzfs_exporter/internal/pool"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// webListener - web server to access the metrics
func (app *application) webListener() {
	log.Printf("listening on %+q", app.listenAddress)
	http.Handle("/metrics", promhttp.Handler())
	if err := app.server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

// application is the settings and state holding structure of the exporter
type application struct {
	server           *http.Server
	interval         time.Duration
	listenAddress    string
	useAutodiscovery bool
	exportedPools    arrayFlags

	poolFilter    *regexp.Regexp
	reverseFilter bool
}

// generate the main application instance
var app = application{}

func init() {
	var poolMatchRaw string
	// parse command line flags
	flag.DurationVar(&app.interval, "interval", 5*time.Second, "refresh interval for metrics")
	flag.StringVar(&app.listenAddress, "web.listen-address", ":8080", "address listening on")
	flag.BoolVar(&app.useAutodiscovery, "discover-pools", false, "use autodiscovery for zfs pools")
	flag.Var(&app.exportedPools, "exported-pools", "address listening on")

	flag.StringVar(&poolMatchRaw, "filter", `^.*$`, "filter queried datasets")
	flag.BoolVar(&app.reverseFilter, "filter-reverse", false, "reverse filter functionality; if set, only not matching datasets would be exported")
	flag.Parse()

	// create http server and assign address
	app.server = &http.Server{
		Handler: nil,
		Addr:    app.listenAddress,
	}

	// autodiscovery mode; append auto discovered hosts to `app.exportedPools`
	if app.useAutodiscovery {
		pools, err := pool.Discover()
		if err != nil {
			log.Fatal(err)
		}
		for _, p := range pools {
			app.exportedPools = append(app.exportedPools, p.Name)
		}
	}

	// only start exporter if pools are set
	if len(app.exportedPools) == 0 {
		log.Fatalln("no pools to check")
	}

	var err error
	if app.poolFilter, err = regexp.Compile(poolMatchRaw); err != nil {
		log.Fatalf("invalid filter expression %+q", err)
	}
}

func main() {
	// capture input signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	ctx, cancelWorkers := context.WithCancel(context.Background())
	done := make(chan interface{})

	// start http listener
	go app.webListener()

	startedWorkers := len(app.exportedPools) - 1
	stoppedWorkers := 0

	// start one worker per pool
	for _, p := range app.exportedPools {
		log.Printf("monitoring p %+q\n", p)
		go app.refreshWorker(ctx, done, p)
	}

	// wait for incoming interrupts
	<-sigChan

	cancelWorkers()

	// wait for all workers to be stopped
	for range done {
		if stoppedWorkers == startedWorkers {
			break
		}
		stoppedWorkers++
	}

	// stop the http service
	if err := app.server.Shutdown(context.Background()); err != nil {
		log.Println(err)
	}
}
