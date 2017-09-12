package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	pge "github.com/crunchydata/postgres_exporter/exporter"
)

var (
	listenAddress = flag.String(
		"web.listen-address", ":9187",
		"Address to listen on for web interface and telemetry.",
	)
	metricPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose metrics.",
	)
	queriesPath = flag.String(
		"extend.query-path", "",
		"Path to custom queries to run.",
	)
	onlyDumpMaps = flag.Bool(
		"dumpmaps", false,
		"Do not run, simply dump the maps.",
	)
	showVersion = flag.Bool("version", false, "print version and exit")
)

// landingPage contains the HTML served at '/'.
// TODO: Make cu nicer and more informative.
var landingPage = []byte(`<html>
<head><title>Postgres exporter</title></head>
<body>
<h1>Postgres exporter</h1>
<p><a href='` + *metricPath + `'>Metrics</a></p>
</body>
</html>
`)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf(
			"postgres_exporter %s (built with %s)\n",
			pge.Version, runtime.Version(),
		)
		return
	}

	if *onlyDumpMaps {
		pge.DumpMaps()
		return
	}

	dsn := os.Getenv("DATA_SOURCE_NAME")
	if len(dsn) == 0 {
		log.Fatal("couldn't find environment variable DATA_SOURCE_NAME")
	}

	exporter := pge.NewExporter(dsn, *queriesPath)
	defer func() {
		conn := exporter.GetConnection()
		if conn != nil {
			conn.Close() // nolint: errcheck
		}
	}()

	prometheus.MustRegister(exporter)

	http.Handle(*metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(landingPage) // nolint: errcheck
	})

	log.Infof("Starting Server: %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
