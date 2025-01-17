package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.APIMetrics {
		ch <- m
	}
}

// Collect function, called on by Prometheus Client library
// This function is called when a scrape is performed on the /metrics page
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	// Scrape the Data from Github
	var data, rates, err = e.gatherData()

	if err != nil {
		log.Errorf("Error gathering Data from remote API: %v", err)
		return
	}

	// Set prometheus gauge metrics using the data gathered
	e.processMetrics(data, rates, ch)

	log.Info("All Metrics successfully collected")

}
