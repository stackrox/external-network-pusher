package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const httpGetTimeout = 60 * time.Second

// HTTPGetWithRetry returns the body of the HTTP Get response. Retries if the call fails.
func HTTPGetWithRetry(provider, url string) ([]byte, error) {
	var body []byte
	retryErr := WithDefaultRetry(func() error {
		var err error
		body, err = HTTPGet(url)
		if err != nil {
			return errors.Wrapf(err, "failed to fetch networks from %s with URL: %s", provider, url)
		}
		return nil
	})
	if retryErr != nil {
		return nil, retryErr
	}
	return body, nil
}

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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non 200 status code. Code: %d, error: %v", resp.StatusCode, err)
	}

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed while trying to copy response data")
	}
	return bodyData, nil
}
