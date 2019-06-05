package main

import (
	"fmt"
	"net/http"

	"github.com/fatih/structs"
	"github.com/guusvw/github-exporter/config"
	"github.com/guusvw/github-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	log            *logrus.Logger
	applicationCfg config.Config
	mets           map[string]*prometheus.Desc
)

const (
	htmlTemp = `<html>
	<head><title>Github Exporter</title></head>
	<body>
	   <h1>GitHub Prometheus Metrics Exporter</h1>
	   <p>For more information, visit <a href=https://github.com/guusvw/github-exporter>GitHub</a></p>
	   <p><a href='%s'>Metrics</a></p>
	   </body>
	</html>
  `
)

func main() {
	applicationCfg = config.Init()
	mets = exporter.AddMetrics()

	log.WithFields(structs.Map(applicationCfg)).Info("Starting Exporter")

	exporter := exporter.Exporter{
		APIMetrics: mets,
		Config:     applicationCfg,
	}

	// Register Metrics from each of the endpoints
	// This invokes the Collect method through the prometheus client libraries.
	prometheus.MustRegister(&exporter)

	// Setup HTTP handler
	http.Handle(applicationCfg.MetricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, htmlTemp, applicationCfg.MetricsPath)
	})
	log.Fatal(http.ListenAndServe(":"+applicationCfg.ListenPort, nil))
}
