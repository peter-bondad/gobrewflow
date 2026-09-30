package paymongo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL   string
	secretKey string
	http      *http.Client
}

func NewClient(baseURL, secretKey string, httpClient *http.Client) *Client {
	// if no http client provided, 10 seconds timeout
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		secretKey: secretKey,
		http:      httpClient,
	}
}

func (c *Client) do(
	ctx context.Context,
	method string,
	path string,
	body []byte,
) ([]byte, error) {

	fmt.Println(string(body))

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		c.baseURL+path,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.secretKey, "")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"paymongo: unexpected status %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	return responseBody, nil
}
