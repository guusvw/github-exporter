package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"

	"os"
)

// Config struct holds all of the runtime confgiguration for the application
type Config struct {
	APIURL        string
	Repositories  string
	Organisations string
	Users         string
	APITokenEnv   string
	APITokenFile  string
	APIToken      string
	MetricsPath   string
	ListenPort    string
	TargetURLs    []string
}

// Init populates the Config struct based on environmental runtime configuration
func Init() Config {

	url := "https://api.github.com"
	repos := os.Getenv("REPOS")
	orgs := os.Getenv("ORGS")
	users := os.Getenv("USERS")
	tokenEnv := os.Getenv("GITHUB_TOKEN")
	tokenFile := os.Getenv("GITHUB_TOKEN_FILE")
	token, err := getAuth(tokenEnv, tokenFile)

	if err != nil {
		log.Errorf("Error initialising Configuration, Error: %v", err)
	}

	scraped, err := getScrapeURLs(url, repos, orgs, users)

	if err != nil {
		log.Errorf("Error initialising Configuration, Error: %v", err)
	}

	return Config{
		APIURL:        url,
		Repositories:  repos,
		Organisations: orgs,
		Users:         users,
		APITokenEnv:   tokenEnv,
		APITokenFile:  tokenFile,
		APIToken:      token,
		TargetURLs:    scraped,
	}
}

// Init populates the Config struct based on environmental runtime configuration
// All URL's are added to the TargetURL's string array
func getScrapeURLs(apiURL, repos, orgs, users string) ([]string, error) {

	urls := []string{}

	opts := "?&per_page=100" // Used to set the Github API to return 100 results per page (max)

	// User input validation, check that either repositories or organisations have been passed in
	if len(repos) == 0 && len(orgs) == 0 && len(users) == 0 {
		return urls, fmt.Errorf("No targets specified")
	}

	// Append repositories to the array
	rs := strings.Split(repos, ", ")
	for _, x := range rs {
		y := fmt.Sprintf("%s/repos/%s%s", apiURL, x, opts)
		urls = append(urls, y)
	}

	// Append github orginisations to the array
	o := strings.Split(orgs, ", ")
	for _, x := range o {
		y := fmt.Sprintf("%s/orgs/%s/repos%s", apiURL, x, opts)
		urls = append(urls, y)
	}

	us := strings.Split(users, ", ")
	for _, x := range us {
		y := fmt.Sprintf("%s/users/%s/repos%s", apiURL, x, opts)
		urls = append(urls, y)
	}

	return urls, nil
}

// getAuth returns oauth2 token as string for usage in http.request
func getAuth(token string, tokenFile string) (string, error) {

	if token != "" {
		return token, nil
	}
	if tokenFile != "" {
		b, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), err
	}

	return "", errors.New("")
}
