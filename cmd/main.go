package main

import (
	"fmt"
	"net/http"

	"github.com/ezhische/qrator-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func healthz(response http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(response, "ok")
}

func main() {
	log := logrus.New()

	conf, err := config.ConfigFromEnv()
	if err != nil {
		log.Fatalf("Can't create config: %v", err)
	}
	coll, err := config.CollectorFromConfig(conf, log)
	if err != nil {
		log.Fatalf("Can't create collector: %v", err)
	}
	prometheus.MustRegister(coll)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Qrator Exporter</title></head>
			<body>
			<h1>Qrator Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})
	log.Infoln("Starting qrator-exporter")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", conf.Port), nil))
}
