package exporter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tomnomnom/linkheader"
)

func asyncHTTPGets(targets []string, token string) []*Response {
	// Expand targets by following GitHub pagination links
	targets = paginateTargets(targets, token)

	// Channels used to enable concurrent requests
	ch := make(chan *Response, len(targets))

	responses := []*Response{}

	for _, url := range targets {

		go func(url string) {
			err := getResponse(url, token, ch)
			if err != nil {
				ch <- &Response{url, nil, []byte{}, err}
			}
		}(url)

	}

	for r := range ch {
		if r.err != nil {
			log.Errorf("Error scraping API, Error: %v", r.err)
			break
		}
		responses = append(responses, r)

	}

	return responses
}

// paginateTargets returns all pages for the provided targets
func paginateTargets(targets []string, token string) []string {

	paginated := targets

	for _, url := range targets {

		// make a request to the original target to get link header if it exists
		resp, err := getHTTPResponse(url, token)
		if err != nil {
			log.Errorf("Error retrieving Link headers, Error: %s", err)
		}

		if resp.Header["Link"] != nil {
			links := linkheader.Parse(resp.Header["Link"][0])

			for _, link := range links {
				if link.Rel == "last" {

					subs := strings.Split(link.URL, "&page=")

					lastPage, err := strconv.Atoi(subs[len(subs)-1])
					if err != nil {
						log.Errorf("Unable to convert page substring to int, Error: %s", err)
					}

					// add all pages to the slice of targets to return
					for page := 2; page <= lastPage; page++ {
						pageURL := fmt.Sprintf("%s&page=%v", url, page)
						paginated = append(paginated, pageURL)
					}

					break
				}
			}
		}
	}
	return paginated
}

// getResponse collects an individual http.response and returns a *Response
func getResponse(url string, token string, ch chan<- *Response) error {

	log.Infof("Fetching %s \n", url)

	resp, err := getHTTPResponse(url, token) // do this earlier

	if err != nil {
		return fmt.Errorf("error converting body to byte array: %v", err)
	}

	// Read the body to a byte array so it can be used elsewhere
	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		return fmt.Errorf("error converting body to byte array: %v", err)
	}

	// Triggers if a user specifies an invalid or not visible repository
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("error: Received 404 status from Github API, ensure the repsository URL is correct. If it's a privare repository, also check the oauth token is correct")
	}

	ch <- &Response{url, resp, body, err}

	return nil
}

// getHTTPResponse handles the http client creation, token setting and returns the *http.response
func getHTTPResponse(url string, token string) (*http.Response, error) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	// If a token is present, add it to the http.request
	if token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", token))
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, err
}
