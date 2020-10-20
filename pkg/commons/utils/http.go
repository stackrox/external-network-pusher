package utils

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const httpGetTimeout = 60 * time.Second

// HTTPGet returns the body of the HTTP GET response
func HTTPGet(url string) ([]byte, error) {
	log.Printf("Getting from URL: %s...", url)

	client := &http.Client{
		Timeout: httpGetTimeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error while copying response data: %+v", err)
		return nil, err
	}
	return bodyData, nil
}