package utilities

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultHTTPClient struct {
	client *http.Client
}

func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 10,
		},
	}
}

func (c *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
