package main

import (
	"flag"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	portptr := flag.String("port", "8080", "Provide the port to listen on")
	flag.Parse()
	port := ":" + *portptr
	//Create a new instance of the foocollector and
	//register it with the prometheus client.
	col := newLxdCollector()
	prometheus.MustRegister(col)

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port ", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
