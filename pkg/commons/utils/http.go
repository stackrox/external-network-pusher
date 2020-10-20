package utils

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const HttpGetTimeout = 60 * time.Second

// This function returns the body of the HTTP GET response
func HttpGet(url string) ([]byte, error) {
	log.Printf("Getting from URL: %s...", url)

	client := &http.Client{
		Timeout: HttpGetTimeout,
	}

	resp, err := client.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error while copying response data: %+v", err)
		return nil, err
	}
	return bodyData, nil
}