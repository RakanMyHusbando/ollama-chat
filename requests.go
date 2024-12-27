package main

import (
	"fmt"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func makeReq(url string) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
