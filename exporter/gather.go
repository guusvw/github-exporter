package exporter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// gatherData - Collects the data from the API and stores into struct
func (e *Exporter) gatherData() ([]*Datum, *RateLimits, error) {

	data := []*Datum{}

	responses := asyncHTTPGets(e.TargetURLs, e.APIToken)

	for _, response := range responses {

		// Github can at times present an array, or an object for the same data set.
		// This code checks handles this variation.
		if isArray(response.body) {
			ds := []*Datum{}
			if err := json.Unmarshal(response.body, &ds); err != nil {
				return nil, nil, err
			}
			data = append(data, ds...)
		} else {
			d := new(Datum)
			if err := json.Unmarshal(response.body, &d); err != nil {
				return nil, nil, err
			}
			data = append(data, d)
		}

		log.Infof("API data fetched for repository: %s", response.url)
	}

	// Check the API rate data and store as a metric
	rates, err := getRates(e.APIURL, e.APIToken)

	if err != nil {
		log.Errorf("Unable to obtain rate limit data from API, Error: %s", err)
		return nil, nil, errors.New("")
	}

	//return data, rates, err
	return data, rates, nil

}

// getRates obtains the rate limit data for requests against the github API.
// Especially useful when operating without oauth and the subsequent lower cap.
func getRates(baseURL string, token string) (*RateLimits, error) {

	rateEndPoint := ("/rate_limit")
	url := fmt.Sprintf("%s%s", baseURL, rateEndPoint)

	resp, err := getHTTPResponse(url, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Triggers if rate-limiting isn't enabled on private Github Enterprise installations
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("rate Limiting not enabled in GitHub API")
	}

	limit, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Limit"), 64)

	if err != nil {
		return nil, err
	}

	rem, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Remaining"), 64)

	if err != nil {
		return nil, err
	}

	reset, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Reset"), 64)

	if err != nil {
		return nil, err
	}

	return &RateLimits{
		Limit:     limit,
		Remaining: rem,
		Reset:     reset,
	}, err

}

// isArray simply looks for key details that determine if the JSON response is an array or not.
func isArray(body []byte) bool {

	isArray := false

	for _, c := range body {
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			continue
		}
		isArray = c == '['
		break
	}

	return isArray

}
